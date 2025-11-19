package workers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
)

type DistributedNode struct {
	raft    *raft.Raft
	fsm     *ClusterState
	config  NodeConfig
	mysqlDB *sql.DB
	ctx     context.Context
	cancel  context.CancelFunc
}

type NodeConfig struct {
	NodeID    string
	ClusterID string
	RaftAddr  string
	HTTPAddr  string
	MySQLDSN  string
	DataDir   string
	Bootstrap bool
}

// TaskRouter 跨集群任务路由
type TaskRouter struct {
	clusterManagers map[string]*MultiClusterManager
}

func NewDistributedNode(cfg NodeConfig) (*DistributedNode, error) {
	// 连接MySQL
	db, err := sql.Open("mysql", cfg.MySQLDSN)
	if err != nil {
		return nil, fmt.Errorf("mysql connect error: %v", err)
	}
	db.SetMaxOpenConns(20)
	db.SetConnMaxLifetime(time.Duration(DefaultDBConnMaxLifetime) * time.Minute)

	// 创建FSM
	fsm := NewClusterState(cfg.ClusterID)

	// Raft配置
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.NodeID)
	raftConfig.Logger = hclog.New(&hclog.LoggerOptions{
		Name:       fmt.Sprintf("RAFT-%s", cfg.NodeID),
		Output:     os.Stderr,
		Level:      hclog.Info,
		TimeFormat: "2006-01-02 15:04:05",
	})
	raftConfig.SnapshotInterval = time.Duration(DefaultRaftSnapshotInterval) * time.Second
	raftConfig.SnapshotThreshold = 100

	// 存储配置
	baseDir := filepath.Join(cfg.DataDir, cfg.ClusterID, cfg.NodeID)
	if err = os.MkdirAll(baseDir, 0755); err != nil {
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
		raft:    r,
		fsm:     fsm,
		config:  cfg,
		mysqlDB: db,
		ctx:     ctx,
		cancel:  cancel,
	}

	// 初始节点引导集群
	if cfg.Bootstrap {
		config := raft.Configuration{
			Servers: []raft.Server{
				{ID: raft.ServerID(cfg.NodeID), Address: transport.LocalAddr()},
			},
		}
		r.BootstrapCluster(config)
		log.Printf("Node %s bootstrapped cluster %s", cfg.NodeID, cfg.ClusterID)
	}

	// 启动后台协程
	go node.leaderHeartbeat()
	go node.monitorState()

	return node, nil
}

// leaderHeartbeat Leader定时更新MySQL中的选主状态
func (n *DistributedNode) leaderHeartbeat() {
	ticker := time.NewTicker(time.Duration(DefaultMasterElectionInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if n.raft.State() == raft.Leader {
				if err := n.updateLeaderInfoInMySQL(); err != nil {
					log.Printf("Failed to update leader info: %v", err)
				}
			}
		case <-n.ctx.Done():
			return
		}
	}
}

// updateLeaderInfoInMySQL 更新Leader信息到MySQL
func (n *DistributedNode) updateLeaderInfoInMySQL() error {
	term := n.raft.CurrentTerm()

	sql := `INSERT INTO leader_election (cluster_id, leader_id, leader_addr, raft_term) 
            VALUES (?, ?, ?, ?) 
            ON DUPLICATE KEY UPDATE 
            leader_id = VALUES(leader_id), 
            leader_addr = VALUES(leader_addr), 
            raft_term = VALUES(raft_term), 
            last_heartbeat = CURRENT_TIMESTAMP`

	_, err := n.mysqlDB.Exec(sql, n.config.ClusterID, n.config.NodeID, n.config.HTTPAddr, term)
	if err != nil {
		log.Printf("Failed to update leader info: %v", err)
		return err
	}

	log.Printf("Leader heartbeat updated: cluster=%s, leader=%s, term=%d",
		n.config.ClusterID, n.config.NodeID, term)
	return nil
}

// monitorState 监控Raft状态变化
func (n *DistributedNode) monitorState() {
	for {
		select {
		case <-n.ctx.Done():
			return
		default:
			if n.raft.State() == raft.Leader {
				n.RunProducer()
			}
			time.Sleep(time.Duration(DefaultWorkerKeepAliveInterval) * time.Second)
		}
	}
}

// Shutdown 优雅关闭
func (n *DistributedNode) Shutdown() error {
	n.cancel()
	future := n.raft.Shutdown()
	err := future.Error()
	// 关闭MySQL连接
	if closeErr := n.mysqlDB.Close(); closeErr != nil {
		log.Printf("Failed to close MySQL connection: %v", closeErr)
		if err == nil {
			err = closeErr
		}
	}

	return err
}

// SyncLeaderFromDB 从MySQL同步Leader信息（供非Leader验证）
func (n *DistributedNode) SyncLeaderFromDB() error {
	query := `SELECT leader_id, leader_addr, raft_term FROM leader_election WHERE cluster_id = ?`
	var leaderID, leaderAddr string
	var term uint64

	// 使用DefaultMasterTimeout作为查询超时
	ctx, cancel := context.WithTimeout(n.ctx, time.Duration(DefaultMasterTimeout)*time.Second)
	defer cancel()

	err := n.mysqlDB.QueryRowContext(ctx, query, n.config.ClusterID).Scan(&leaderID, &leaderAddr, &term)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // 无记录
		}
		return err
	}

	// 验证本地Raft状态与数据库是否一致
	if n.fsm.leaderID != leaderID || n.fsm.currentTerm != term {
		log.Printf("Leader info mismatch: DB(%s, term:%d) vs Local(%s, term:%d)",
			leaderID, term, n.fsm.leaderID, n.fsm.currentTerm)
	}

	return nil
}

// GetLeader 获取当前Leader节点ID
func (n *DistributedNode) GetLeader() string {
	n.fsm.mu.RLock()
	defer n.fsm.mu.RUnlock()
	return n.fsm.leaderID
}

// IsLeader 检查当前节点是否为Leader
func (n *DistributedNode) IsLeader() bool {
	return n.raft.State() == raft.Leader
}

// RunProducer 运行任务生产者
func (n *DistributedNode) RunProducer() {
	// 创建TaskProducer实例
	producer := NewTaskProducer(n)
	producer.RunProducer()
}

// HandleMultiCluster 多集群管理
type MultiClusterManager struct {
	clusters map[string]*DistributedNode
}

func (m *MultiClusterManager) StartCluster(clusterID string, cfg NodeConfig) error {
	if _, exists := m.clusters[clusterID]; exists {
		return fmt.Errorf("cluster %s already exists", clusterID)
	}

	node, err := NewDistributedNode(cfg)
	if err != nil {
		return err
	}

	m.clusters[clusterID] = node
	return nil
}

func (m *MultiClusterManager) GetLeader(clusterID string) string {
	if node, exists := m.clusters[clusterID]; exists {
		return node.GetLeader()
	}
	return ""
}

// RouteTask 根据task类型路由到不同集群
func (r *TaskRouter) RouteTask(taskType string, payload interface{}) error {
	// 配置规则：email任务 -> cluster1, data任务 -> cluster2
	clusterMap := map[string]string{
		"email_send":    "cluster1",
		"data_sync":     "cluster2",
		"image_process": "cluster3",
	}

	clusterID, ok := clusterMap[taskType]
	if !ok {
		return fmt.Errorf("unknown task type: %s", taskType)
	}

	// 获取目标集群的Leader
	leader := r.clusterManagers[clusterID].GetLeader(clusterID)
	if leader == "" {
		return fmt.Errorf("cluster %s has no leader", clusterID)
	}

	// 将任务写入目标集群的MySQL
	// ...
	return nil
}
