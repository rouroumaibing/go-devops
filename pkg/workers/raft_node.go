package workers

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

// 添加包级变量用于控制leader心跳日志频率
var leaderHeartbeatLogCounter = 0
var leaderHeartbeatLogInterval = 30 // 每30次心跳才打印一次成功日志

type NodeConfig struct {
	RaftClusterID  string // 全局raft集群id
	NodeIP         string // 节点IP地址
	ClusterID      string // 子集群id
	CombinedNodeID string // 节点ID, 格式为'ClusterID@NodeIP'
	RaftAddr       string // 节点Raft监听地址
	DataDir        string // 节点数据目录
	MaxIdleConns   int    // 数据库最大空闲连接数
	MaxOpenConns   int    // 数据库最大打开连接数
}

type DistributedNode struct {
	raft   *raft.Raft
	fsm    *ClusterFSM
	config NodeConfig
	ctx    context.Context
	cancel context.CancelFunc
}

func NewDistributedNode(cfg NodeConfig) (*DistributedNode, error) {
	// 都放在大集群RaftClusterID下，集群节点变为'ClusterID@NodeIP'
	fsm := NewClusterFSM(cfg.RaftClusterID)

	combinedID := fmt.Sprintf("%s@%s", cfg.ClusterID, cfg.NodeIP)
	cfg.CombinedNodeID = combinedID

	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.CombinedNodeID)
	raftConfig.Logger = hclog.New(&hclog.LoggerOptions{
		Name:       fmt.Sprintf("RAFT-%s", cfg.CombinedNodeID),
		Output:     os.Stderr,
		Level:      hclog.Info,
		TimeFormat: "2006-01-02 15:04:05",
	})
	raftConfig.SnapshotInterval = time.Duration(DefaultRaftSnapshotInterval) * time.Second
	raftConfig.SnapshotThreshold = 100

	// 检查leader_election的更新时间，如果超过5分钟,说明数据库集群信息无效，可以重新初始化
	CheckLeaderElectionRecord(&cfg)

	// 存储配置
	baseDir := filepath.Join(cfg.DataDir, cfg.RaftClusterID, cfg.NodeIP)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(baseDir, "raft-log.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create log store: %v", err)
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(baseDir, "raft-stable.db"))
	if err != nil {
		return nil, fmt.Errorf("failed to create stable store: %v", err)
	}
	snapshotStore, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot store: %v", err)
	}
	// 传输层
	addr, err := net.ResolveTCPAddr("tcp", cfg.RaftAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve raft address: %v", err)
	}
	transport, err := raft.NewTCPTransport(cfg.RaftAddr, addr, 5, time.Duration(DefaultRaftTransportTimeout)*time.Second, os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("failed to create transport: %v", err)
	}

	// 创建Raft实例
	r, err := raft.NewRaft(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	node := &DistributedNode{
		raft:   r,
		fsm:    fsm,
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
	}

	// 自动判断是否需要引导集群
	// 1. 检查数据库中是否已经存在集群信息
	var exists bool
	var count int
	dao := orm.NewOrm()
	err = dao.Raw("SELECT COUNT(1) FROM leader_election WHERE leader_id = ?", cfg.CombinedNodeID).QueryRow(&count)
	exists = count > 0
	if err != nil {
		log.Printf("Warning: failed to check if cluster exists in database: %v", err)
	}

	// 2. 检查Raft数据目录是否已经存在
	raftDirExists := false
	if _, err = os.Stat(filepath.Join(baseDir, "raft")); err == nil {
		raftDirExists = true
	}

	// 如果数据库中没有集群信息且Raft目录不存在，则引导集群
	if !exists && !raftDirExists {
		config := raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(cfg.CombinedNodeID), Address: transport.LocalAddr()},
			},
		}
		r.BootstrapCluster(config)
		log.Printf("Node %s bootstrapped cluster %s (auto-detected as first node)", cfg.NodeIP, cfg.ClusterID)
	} else if exists {
		log.Printf("Node %s joining existing cluster %s (found in database)", cfg.NodeIP, cfg.ClusterID)
		go node.AddNodeToJoinClusterList(&cfg)
	} else {
		log.Printf("Node %s joining existing cluster %s (found local raft data)", cfg.NodeIP, cfg.ClusterID)
	}
	go node.leaderHeartbeat()
	go node.NodeHealthCheck()
	go node.selfCheckLoop()

	return node, nil
}

func CheckLeaderElectionRecord(cfg *NodeConfig) {
	dao := orm.NewOrm()
	// 检查leader_election记录的更新时间是否超过5分钟
	// 如果超过5分钟没有更新，说明raft集群整个信息都是无效的，需要重新初始化
	// updated_at > NOW() - INTERVAL 5 MINUTE:
	//     小于等于5分钟的数据不存在,均返回错误<QuerySeter> no row found
	//     满足条件，会查询到1条记录，返回值1，不会报错
	var isFresh int
	err := dao.Raw("SELECT 1 FROM leader_election WHERE raft_clusterid = ? AND updated_at > NOW() - INTERVAL 5 MINUTE", cfg.RaftClusterID).QueryRow(&isFresh)
	if err != nil {
		log.Printf("Warning: leader election record is stale (older than 5 minutes):  %s", cfg.RaftClusterID)
		// 清理leader_election记录
		_, delErr := dao.Raw("DELETE FROM leader_election WHERE raft_clusterid = ?", cfg.RaftClusterID).Exec()
		if delErr != nil {
			log.Printf("Error: failed to delete stale leader election record: %v", delErr)
		} else {
			log.Printf("Deleted stale leader election record for cluster %s", cfg.RaftClusterID)
		}
		// 清理本地data目录
		cleanErr := os.RemoveAll(filepath.Join(DefaultDataDir))
		if cleanErr != nil {
			log.Printf("Error: failed to remove data directory: %v", cleanErr)
		} else {
			log.Printf("Cleaned local data directory for cluster %s", cfg.RaftClusterID)
		}
	}

}

func (n *DistributedNode) GetNodeIP() string {
	return n.config.NodeIP
}

// 节点ID (clusterID@nodeIP)
func (n *DistributedNode) GetCombinedNodeID() string {
	return n.config.CombinedNodeID
}

func GetNodeIPFromCombinedID(combinedNodeID string) string {
	parts := strings.Split(combinedNodeID, "@")
	if len(parts) != 2 {
		return ""
	}
	return parts[1]
}

func (n *DistributedNode) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

// 获取Leader节点IP
func (n *DistributedNode) GetLeaderIP() string {
	n.fsm.mu.RLock()
	defer n.fsm.mu.RUnlock()

	return n.fsm.leaderIP
}

func (n *DistributedNode) leaderHeartbeat() {

	ticker := time.NewTicker(time.Duration(DefaultMasterElectionInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if n.raft.State() == raft.Leader {
				leaderHeartbeatLogCounter++
				err := n.updateLeaderToDB(leaderHeartbeatLogCounter)
				if err != nil {
					// 错误日志仍然每次都打印
					log.Printf("Failed to update leader info: %v", err)
				}
			}
		case <-n.ctx.Done():
			return
		}
	}
}

func (n *DistributedNode) updateLeaderToDB(logCounter int) error {

	term := n.raft.CurrentTerm()

	var count int
	dao := orm.NewOrm()
	_ = dao.Raw("SELECT COUNT(1) FROM leader_election WHERE raft_clusterid = ?", n.config.RaftClusterID).QueryRow(&count)
	if count > 0 {
		sql := `UPDATE leader_election SET leader_id= ?, leader_ip = ?, raft_term = ?, leader_addr = ?, last_heartbeat = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
            WHERE raft_clusterid = ?`
		_, err := dao.Raw(sql, n.config.CombinedNodeID, n.config.NodeIP, term, n.config.RaftAddr, n.config.RaftClusterID).Exec()
		if err != nil {
			log.Printf("Failed to update leader info: %v", err)
			return err
		}
	} else {
		// 第一次选主成功，插入leader选主信息
		newUUID := uuid.New().String()
		sql := `INSERT INTO leader_election (id, leader_id, leader_ip, raft_clusterid, raft_term, leader_addr, last_heartbeat, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		_, err := dao.Raw(sql, newUUID, n.config.CombinedNodeID, n.config.NodeIP, n.config.RaftClusterID, term, n.config.RaftAddr).Exec()
		if err != nil {
			log.Printf("Failed to insert leader info: %v", err)
			return err
		}
	}

	// 仅当需要打印日志时才输出
	if logCounter%leaderHeartbeatLogInterval == 0 {
		log.Printf("Leader heartbeat updated: cluster=%s, leader=%s, term=%d",
			n.config.ClusterID, n.config.CombinedNodeID, term)
		leaderHeartbeatLogCounter = 0 // 重置计数器
	}
	return nil
}

// Shutdown 优雅关闭
func (n *DistributedNode) Shutdown() error {
	n.cancel()
	future := n.raft.Shutdown()
	return future.Error()
}

func (n *DistributedNode) JoinCluster(ncfg *NodeConfig) error {
	// 只有Leader可以添加新节点到集群
	if !n.IsLeader() {
		return fmt.Errorf("only leader can add nodes to cluster")
	}

	// 解析新节点的地址
	tcpAddr, err := net.ResolveTCPAddr("tcp", ncfg.RaftAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	// 使用raft.AddVoter添加新节点
	log.Printf("Adding server %s (raft cluster name: %s) at %s to cluster", ncfg.CombinedNodeID, ncfg.RaftClusterID, ncfg.RaftAddr)
	future := n.raft.AddVoter(raft.ServerID(ncfg.RaftClusterID), raft.ServerAddress(tcpAddr.String()), 0, 0)
	if err := future.Error(); err != nil {
		return fmt.Errorf("failed to join cluster: %v", err)
	}

	log.Printf("Server %s (raft cluster name: %s) added to cluster successfully", ncfg.CombinedNodeID, ncfg.RaftClusterID)
	return nil
}

// 将节点添加到加入集群队列
func (n *DistributedNode) AddNodeToJoinClusterList(ncfg *NodeConfig) {
	// 检查节点自身是否是Leader,如果是,不处理
	if n.IsLeader() {
		log.Printf("Node %s is already a leader, no need to register", n.config.NodeIP)
		return
	}
	// 检查节点是否已经在集群中
	exists, err := n.CheckNodeExists(ncfg)
	if err != nil {
		log.Printf("Failed to check node existence: %v", err)
		return
	}
	// 如果节点已经在集群中,不处理
	if exists {
		log.Printf("Node %s joining existing cluster %s (found in database)", ncfg.NodeIP, ncfg.ClusterID)
		return
	}
	// 如果节点不存在,加入到待加入集群列表
	JoinClusterList = append(JoinClusterList, ncfg)

}

// 主节点处理加入集群列表中待加入节点
func (n *DistributedNode) RegisterNodeFromJoinClusterList() {
	if !n.IsLeader() {
		log.Printf("Node %s is not a leader, no need to register", n.config.NodeIP)
		return
	}

	// 使用重试函数进行注册
	err := retryOperation(func() error {
		dao := orm.NewOrm()
		// 从数据库获取当前Leader信息
		query := `SELECT leader_addr FROM leader_election WHERE raft_clusterid = ? ORDER BY last_heartbeat DESC LIMIT 1`
		var leaderAddr string
		err := dao.Raw(query, n.config.RaftClusterID).QueryRow(&leaderAddr)
		if err != nil {
			return fmt.Errorf("failed to get leader information: %v", err)
		}
		// 创建Leader节点的临时连接
		leaderConn, err := net.Dial("tcp", leaderAddr)
		if err != nil {
			return fmt.Errorf("failed to connect to leader: %v", err)
		}
		leaderConn.Close() // 只检查连接性，不需要保持连接

		// 从待加入集群列表中循环获取节点信息
		for _, node := range JoinClusterList {
			err = n.JoinCluster(node)
			if err != nil {
				return fmt.Errorf("failed to join cluster: %v", err)
			}
		}
		return nil
	}, 10, 5*time.Second)

	if err != nil {
		log.Printf("Failed to register node to cluster after 10 attempts: %v", err)
	}
}

func (n *DistributedNode) NodeHealthCheck() {
	ticker := time.NewTicker(DefaultRaftNodeHealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取集群节点列表
			conf := n.raft.GetConfiguration()
			if conf.Error() != nil {
				log.Printf("Failed to get cluster configuration: %v", conf.Error())
				continue
			}

			// 周期性探测，超过多少时间，认为节点已经被销毁，需要加入到删除队列中
			// 遍历所有服务器,检查是否可连接。server.Address默认格式为ip:port
			for _, server := range conf.Configuration().Servers {
				// 跳过自身节点
				if string(server.ID) == n.config.CombinedNodeID {
					continue
				}

				testConn, err := net.Dial("tcp", string(server.Address))
				if err != nil {
					// 继续尝试3次，每次间隔累计
					log.Printf("Failed to connect to node %s (address: %s), retrying...", server.ID, server.Address)
					retrySuccess := false

					// 重试3次，间隔时间逐渐增加 (1s, 2s, 3s)
					for i := 1; i <= 3; i++ {
						// 检查context是否已取消
						select {
						case <-n.ctx.Done():
							log.Printf("Health check loop shutting down during retry on node %s", n.config.NodeIP)
							return
						default:
							// 继续重试
						}

						// 累计间隔时间：第1次1s，第2次4s，第3次9s
						retryDelay := time.Duration(i*i) * time.Second
						log.Printf("Retry attempt %d for node %s, waiting %v...", i+1, server.ID, retryDelay)
						time.Sleep(retryDelay)

						retryConn, retryErr := net.Dial("tcp", string(server.Address))
						if retryErr == nil {
							log.Printf("Successfully connected to node %s on retry attempt %d", server.ID, i+1)
							retryConn.Close()
							retrySuccess = true
							break
						}
						log.Printf("Retry attempt %d for node %s failed: %v", i+1, server.ID, retryErr)
					}

					// 超过3次后，将节点加入到删除队列中
					if !retrySuccess {
						log.Printf("Node %s failed all 3 connection attempts, adding to removal queue", server.ID)
						nodeIP := GetNodeIPFromCombinedID(string(server.ID))
						ncfg := &NodeConfig{
							CombinedNodeID: string(server.ID),
							NodeIP:         nodeIP,
							RaftAddr:       string(server.Address),
							RaftClusterID:  n.config.RaftClusterID,
							ClusterID:      n.config.ClusterID,
						}

						n.AddNodeToRemoveClusterList(ncfg)
						go n.RemoveNodeFromCluster()
					}
				} else {
					log.Printf("Successfully connected to node %s", server.ID)
					testConn.Close()
				}
			}
		case <-n.ctx.Done():
			log.Printf("Health check loop shutting down on node %s", n.config.NodeIP)
			return
		}
	}
}

func (n *DistributedNode) AddNodeToRemoveClusterList(ncfg *NodeConfig) {
	// 检查节点是否已经在删除队列中
	for _, node := range RemoveClusterList {
		if node.CombinedNodeID == ncfg.CombinedNodeID {
			log.Printf("Node %s is already in remove cluster list", ncfg.CombinedNodeID)
			return
		}
	}
	// 如果节点不存在,加入到待删除集群列表
	RemoveClusterList = append(RemoveClusterList, ncfg)
}

// 移除节点
func (n *DistributedNode) RemoveNodeFromCluster() {
	// 只有Leader可以移除节点
	if !n.IsLeader() {
		return
	}
	// 从删除节点队列中获取节点
	for _, node := range RemoveClusterList {
		// 使用raft.RemoveServer移除节点
		log.Printf("Removing server %s  from cluster", node.CombinedNodeID)
		future := n.raft.RemoveServer(raft.ServerID(node.CombinedNodeID), 0, 0)
		if err := future.Error(); err != nil {
			return
		}
	}

	log.Printf("Server %s (combinedID: %s) removed from cluster successfully", n.config.CombinedNodeID, n.config.CombinedNodeID)
}

// 检查节点是否存在
func (n *DistributedNode) CheckNodeExists(ncfg *NodeConfig) (bool, error) {
	conf := n.raft.GetConfiguration()
	if conf.Error() != nil {
		return false, conf.Error()
	}

	// 遍历所有服务器
	for _, server := range conf.Configuration().Servers {
		if string(server.ID) == ncfg.CombinedNodeID {
			return true, nil // 节点已加入
		}
	}
	return false, nil
}

// 重试函数，用于执行需要多次尝试的操作
func retryOperation(operation func() error, maxRetries int, delay time.Duration) error {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if err := operation(); err != nil {
			lastErr = err
			log.Printf("Attempt %d failed: %v, retrying after %v...", i+1, err, delay)
			time.Sleep(delay)
			continue
		}
		return nil // 操作成功
	}
	return lastErr
}

// 在节点上运行自检协程
func (n *DistributedNode) selfCheckLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			conf := n.raft.GetConfiguration()
			if conf.Error() != nil {
				continue
			}

			// 检查自身 ID 是否还在配置中
			found := false
			for _, s := range conf.Configuration().Servers {
				// 使用节点配置中的CombinedNodeID替代raft.LocalID()方法
				if s.ID == raft.ServerID(n.config.CombinedNodeID) {
					found = true
					break
				}
			}

			if !found {
				log.Fatal("Node has been removed from cluster! Shutting down...")
				// 或触发告警: n.alerts.Send("node-removed")
				n.raft.Shutdown()
				os.Exit(1)
			}
		case <-n.ctx.Done():
			log.Printf("Self-check loop shutting down on node %s", n.config.NodeIP)
			return
		}
	}
}
