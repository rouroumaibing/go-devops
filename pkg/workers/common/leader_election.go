package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/rouroumaibing/go-devops/pkg/apis/leaders"
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
	"k8s.io/klog/v2"
)

// LeaderElectionManager 管理选主相关操作
type LeaderElectionManager struct {
	raft        *raft.Raft
	config      *Config
	leaderStore *LeaderStore
	mu          sync.RWMutex
	isLeader    bool
}

// Config 存储选主配置
type Config struct {
	// 节点ID
	NodeID string
	// 节点地址
	Address string
	// 数据目录
	DataDir string
	// 集群地址列表
	Peers []string
	// 集群ID
	ClusterID string
	// 选主类型(master/worker)
	Kind string
	// 选举超时时间
	ElectionTimeout time.Duration
	// 心跳间隔
	HeartbeatTimeout time.Duration
	// 日志应用超时时间
	ApplyTimeout time.Duration
	// 快照保留数量
	SnapshotRetainCount int
	// Leader变更回调函数
	LeaderChangerListener func(isLeader bool)
}

type LeaderStore struct {
	leaderInfo *leaders.LeaderElection
	mu         sync.RWMutex
	orm        orm.Ormer
}

// 注册LeaderElection模型到ORM
func init() {
	orm.RegisterModel(new(leaders.LeaderElection))
}

type FSM struct {
	store *LeaderStore
}

// Apply 应用日志到状态机
func (fsm *FSM) Apply(log *raft.Log) interface{} {
	var cmd map[string]string
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		klog.Errorf("Error unmarshaling command: %v", err)
		return nil
	}

	if cmd["type"] == "leader_change" {
		fsm.store.mu.Lock()
		defer fsm.store.mu.Unlock()

		leaderInfo := &leaders.LeaderElection{
			Id:        cmd["id"],
			Kind:      cmd["kind"],
			ClusterID: cmd["clusterId"],
			LeaderIP:  cmd["leaderIp"],
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		fsm.store.leaderInfo = leaderInfo

		// 保存到数据库
		if err := fsm.store.saveToDB(leaderInfo); err != nil {
			klog.Errorf("Failed to save leader info to database: %v", err)
		}

		return leaderInfo
	}

	return nil
}

func (fsm *FSM) Snapshot() (raft.FSMSnapshot, error) {
	fsm.store.mu.RLock()
	defer fsm.store.mu.RUnlock()

	snapshot := &FSMSnapshot{
		leaderInfo: fsm.store.leaderInfo,
	}

	return snapshot, nil
}

// Restore 从快照恢复状态机
func (fsm *FSM) Restore(snapshot io.ReadCloser) error {
	defer snapshot.Close()

	var leaderInfo leaders.LeaderElection
	if err := json.NewDecoder(snapshot).Decode(&leaderInfo); err != nil {
		return err
	}

	fsm.store.mu.Lock()
	defer fsm.store.mu.Unlock()

	fsm.store.leaderInfo = &leaderInfo

	// 恢复到数据库
	if err := fsm.store.saveToDB(&leaderInfo); err != nil {
		klog.Errorf("Failed to restore leader info to database: %v", err)
		// 继续执行，不阻止恢复
	}

	return nil
}

type FSMSnapshot struct {
	leaderInfo *leaders.LeaderElection
}

func (snapshot *FSMSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		if snapshot.leaderInfo == nil {
			return sink.Close()
		}

		if err := json.NewEncoder(sink).Encode(snapshot.leaderInfo); err != nil {
			return err
		}

		return sink.Close()
	}()

	if err != nil {
		if cancelErr := sink.Cancel(); cancelErr != nil {
			err = fmt.Errorf("%w: %v", err, cancelErr)
		}
	}

	return err
}

// Release 释放快照资源
func (snapshot *FSMSnapshot) Release() {
	// 释放资源
	snapshot.leaderInfo = nil
}

// NewLeaderElectionManager 创建选主管理器
func NewLeaderElectionManager(config *Config) (*LeaderElectionManager, error) {
	// 验证配置
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %v", err)
	}

	// 创建数据目录
	if err := os.MkdirAll(config.DataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %v", err)
	}

	// 设置默认配置值
	if config.ElectionTimeout == 0 {
		config.ElectionTimeout = 10 * time.Second
	}
	if config.HeartbeatTimeout == 0 {
		config.HeartbeatTimeout = 1 * time.Second
	}
	if config.ApplyTimeout == 0 {
		config.ApplyTimeout = 5 * time.Second
	}
	if config.SnapshotRetainCount == 0 {
		config.SnapshotRetainCount = 2
	}

	// 创建LeaderStore并初始化数据库连接
	store := &LeaderStore{
		leaderInfo: nil,
		orm:        orm.NewOrm(),
	}

	// 从数据库加载现有Leader信息
	if err := store.loadFromDB(config.ClusterID); err != nil {
		klog.Warningf("Failed to load leader info from database: %v", err)
		// 继续执行，不阻止初始化
	}

	manager := &LeaderElectionManager{
		config:      config,
		leaderStore: store,
		isLeader:    false,
	}

	if err := manager.initRaft(); err != nil {
		return nil, fmt.Errorf("failed to initialize raft: %v", err)
	}

	return manager, nil
}

// validateConfig 验证配置有效性
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if config.NodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}
	if config.Address == "" {
		return fmt.Errorf("address cannot be empty")
	}
	if config.DataDir == "" {
		return fmt.Errorf("data directory cannot be empty")
	}
	if config.ClusterID == "" {
		return fmt.Errorf("cluster ID cannot be empty")
	}
	if config.Kind == "" {
		return fmt.Errorf("kind cannot be empty")
	}
	return nil
}

// initRaft 初始化raft
func (m *LeaderElectionManager) initRaft() error {
	// 设置Raft配置
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(m.config.NodeID)
	config.ElectionTimeout = m.config.ElectionTimeout
	config.HeartbeatTimeout = m.config.HeartbeatTimeout

	// 设置日志存储
	logStorePath := filepath.Join(m.config.DataDir, "raft.log")
	logStore, err := raftboltdb.NewBoltStore(logStorePath)
	if err != nil {
		return fmt.Errorf("failed to create log store: %v", err)
	}

	// 设置稳定存储
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(m.config.DataDir, "raft.stable"))
	if err != nil {
		logStore.Close()
		return fmt.Errorf("failed to create stable store: %v", err)
	}

	// 设置快照存储
	snapshotStore, err := raft.NewFileSnapshotStore(m.config.DataDir, m.config.SnapshotRetainCount, os.Stderr)
	if err != nil {
		logStore.Close()
		stableStore.Close()
		return fmt.Errorf("failed to create snapshot store: %v", err)
	}

	// 创建网络层
	addr, err := net.ResolveTCPAddr("tcp", m.config.Address)
	if err != nil {
		logStore.Close()
		stableStore.Close()
		return fmt.Errorf("failed to resolve address: %v", err)
	}

	transport, err := raft.NewTCPTransport(m.config.Address, addr, 3, 10*time.Second, os.Stderr)
	if err != nil {
		logStore.Close()
		stableStore.Close()
		return fmt.Errorf("failed to create transport: %v", err)
	}

	// 创建FSM
	fsm := &FSM{
		store: m.leaderStore,
	}

	// 创建Raft实例
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		logStore.Close()
		stableStore.Close()
		transport.Close()
		return fmt.Errorf("failed to create raft: %v", err)
	}

	// 检查是否已有集群状态
	hasExistingState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
	if err != nil {
		r.Shutdown()
		return fmt.Errorf("failed to check existing state: %v", err)
	}

	m.raft = r

	// 如果没有现有状态，初始化为单节点集群
	if !hasExistingState && len(m.config.Peers) == 0 {
		servers := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(m.config.NodeID),
					Address: transport.LocalAddr(),
				},
			},
		}

		future := r.BootstrapCluster(servers)
		if err := future.Error(); err != nil {
			r.Shutdown()
			return fmt.Errorf("failed to bootstrap cluster: %v", err)
		}
		klog.Infof("Bootstrapped new single-node cluster with node ID: %s", m.config.NodeID)
	} else if len(m.config.Peers) > 0 {
		// 尝试加入现有集群
		go m.joinExistingCluster()
	}

	// 监听Leader变更事件
	go m.monitorLeaderChanges()

	return nil
}

// joinExistingCluster 尝试加入现有集群
func (m *LeaderElectionManager) joinExistingCluster() {
	// 延迟加入，给Raft实例一些初始化时间
	time.Sleep(2 * time.Second)

	// 检查是否已经是Leader，如果是则无需加入
	if m.IsLeader() {
		klog.Infof("Node %s is already a leader, no need to join existing cluster", m.config.NodeID)
		return
	}

	for _, peer := range m.config.Peers {
		if peer == m.config.Address {
			continue
		}

		klog.Infof("Attempting to join existing cluster through peer: %s", peer)

		// 注意：使用AddNonvoter而不是AddVoter进行初始加入，这样可以先同步日志再获得投票权
		future := m.raft.AddNonvoter(raft.ServerID(m.config.NodeID), raft.ServerAddress(m.config.Address), 0, 0)
		if err := future.Error(); err != nil {
			klog.Warningf("Failed to add self as non-voter through peer %s: %v", peer, err)
			// 尝试直接添加为voter（可能是旧版本Raft实现不支持non-voter）
			future := m.raft.AddVoter(raft.ServerID(m.config.NodeID), raft.ServerAddress(m.config.Address), 0, 0)
			if err := future.Error(); err != nil {
				klog.Warningf("Failed to add self as voter through peer %s: %v", peer, err)
				// 继续尝试下一个peer
				continue
			}
			klog.Infof("Successfully joined cluster as voter through peer: %s", peer)
		} else {
			// 成功添加为non-voter后，等待一段时间再升级为voter
			klog.Infof("Successfully joined cluster as non-voter through peer: %s", peer)
			// 延迟后升级为voter，给节点时间同步日志
			time.Sleep(3 * time.Second)
			promoteFuture := m.raft.AddVoter(raft.ServerID(m.config.NodeID), raft.ServerAddress(m.config.Address), 0, 0)
			if err := promoteFuture.Error(); err != nil {
				klog.Warningf("Failed to promote self to voter: %v", err)
			} else {
				klog.Infof("Successfully promoted self to voter")
			}
		}

		return
	}

	klog.Errorf("Failed to join any existing cluster peers. Consider checking network connectivity or if the cluster is running.")
}

// monitorLeaderChanges 监听Leader变更事件
func (m *LeaderElectionManager) monitorLeaderChanges() {
	// 定期检查Leader状态
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		isLeader := m.raft.State() == raft.Leader

		m.mu.Lock()
		hasChanged := isLeader != m.isLeader
		previousState := m.isLeader
		m.isLeader = isLeader
		m.mu.Unlock()

		if hasChanged {
			if isLeader {
				klog.Infof("Node %s became leader", m.config.NodeID)
				// 发布Leader变更消息
				m.publishLeaderChange()
			} else {
				klog.Infof("Node %s is no longer leader", m.config.NodeID)
			}

			// 调用Leader变更回调函数
			if m.config.LeaderChangerListener != nil {
				go m.config.LeaderChangerListener(isLeader)
			}

			klog.Infof("Leader state changed: %v -> %v", previousState, isLeader)
		}
	}
}

// publishLeaderChange 发布Leader变更消息
func (m *LeaderElectionManager) publishLeaderChange() {
	// 获取本机IP
	ip, err := system.GetLocalIP()
	if err != nil {
		klog.Errorf("Failed to get local IP for node %s: %v", m.config.NodeID, err)
		return
	}

	// 创建Leader变更命令
	cmd := map[string]string{
		"type":      "leader_change",
		"id":        m.config.NodeID,
		"kind":      m.config.Kind,
		"clusterId": m.config.ClusterID,
		"leaderIp":  ip,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		klog.Errorf("Failed to marshal leader change command for node %s: %v", m.config.NodeID, err)
		return
	}

	// 向Raft提交命令
	future := m.raft.Apply(cmdBytes, m.config.ApplyTimeout)
	if err := future.Error(); err != nil {
		klog.Errorf("Failed to apply leader change command for node %s: %v", m.config.NodeID, err)
		return
	}

	// 记录成功日志，包含更多上下文信息
	result := future.Response()
	klog.Infof("Leader change published successfully for node %s (IP: %s), cluster: %s, result: %v",
		m.config.NodeID, ip, m.config.ClusterID, result)
}

// IsLeader 检查当前节点是否为Leader
func (m *LeaderElectionManager) IsLeader() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isLeader
}

// GetLeader 获取当前Leader信息
func (m *LeaderElectionManager) GetLeader() *leaders.LeaderElection {
	m.leaderStore.mu.RLock()
	defer m.leaderStore.mu.RUnlock()
	return m.leaderStore.leaderInfo
}

// GetLeaderFromDB 从数据库获取Leader信息
func (m *LeaderElectionManager) GetLeaderFromDB() (*leaders.LeaderElection, error) {
	return m.leaderStore.getFromDB(m.config.ClusterID)
}

// saveToDB 保存Leader信息到数据库
func (s *LeaderStore) saveToDB(leaderInfo *leaders.LeaderElection) error {
	if leaderInfo == nil {
		return fmt.Errorf("leaderInfo is nil")
	}

	// 使用事务确保数据一致性
	dao := s.orm
	err := dao.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	// 先删除该集群的旧leader记录
	_, err = dao.Raw("DELETE FROM leader_election WHERE cluster_id = ?", leaderInfo.ClusterID).Exec()
	if err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			klog.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to delete old leader record: %v", err)
	}

	// 插入新的leader记录
	_, err = dao.Insert(leaderInfo)
	if err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			klog.Errorf("failed to rollback transaction: %v", rollbackErr)
		}
		return fmt.Errorf("failed to insert leader record: %v", err)
	}

	// 提交事务
	if err := dao.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	klog.Infof("Successfully saved leader info to database: ID=%s, ClusterID=%s",
		leaderInfo.Id, leaderInfo.ClusterID)
	return nil
}

// getFromDB 从数据库获取Leader信息
func (s *LeaderStore) getFromDB(clusterID string) (*leaders.LeaderElection, error) {
	var leaderInfo leaders.LeaderElection
	err := s.orm.Raw("SELECT * FROM leader_election WHERE cluster_id = ? ORDER BY updated_at DESC LIMIT 1",
		clusterID).QueryRow(&leaderInfo)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, nil // 没有找到记录
		}
		return nil, fmt.Errorf("failed to query leader from database: %v", err)
	}
	return &leaderInfo, nil
}

// loadFromDB 从数据库加载Leader信息到内存
func (s *LeaderStore) loadFromDB(clusterID string) error {
	leaderInfo, err := s.getFromDB(clusterID)
	if err != nil {
		return err
	}

	if leaderInfo != nil {
		s.mu.Lock()
		s.leaderInfo = leaderInfo
		s.mu.Unlock()
		klog.Infof("Successfully loaded leader info from database: ID=%s, ClusterID=%s",
			leaderInfo.Id, leaderInfo.ClusterID)
	}
	return nil
}

// AddPeer 添加集群节点，支持重试机制
func (m *LeaderElectionManager) AddPeer(nodeID, address string) error {
	// 检查是否为Leader
	if !m.IsLeader() {
		return fmt.Errorf("only leader can add peers")
	}

	if nodeID == "" || address == "" {
		return fmt.Errorf("node ID and address cannot be empty")
	}

	// 添加节点到Raft配置，使用重试机制
	maxRetries := 3
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		klog.Infof("Attempting to add peer %s at %s (attempt %d/%d)", nodeID, address, i+1, maxRetries)
		future := m.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(address), 0, 0)
		err := future.Error()
		if err == nil {
			klog.Infof("Successfully added peer %s at %s", nodeID, address)
			return nil
		}

		if i < maxRetries-1 {
			klog.Warningf("Failed to add peer %s at %s, will retry in %v: %v", nodeID, address, retryDelay, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // 指数退避
		}
	}

	return fmt.Errorf("failed to add peer %s at %s after %d attempts", nodeID, address, maxRetries)
}

// RemovePeer 移除集群节点，支持重试机制
func (m *LeaderElectionManager) RemovePeer(nodeID string) error {
	// 检查是否为Leader
	if !m.IsLeader() {
		return fmt.Errorf("only leader can remove peers")
	}

	if nodeID == "" {
		return fmt.Errorf("node ID cannot be empty")
	}

	// 移除节点，使用重试机制
	maxRetries := 3
	retryDelay := 1 * time.Second

	for i := 0; i < maxRetries; i++ {
		klog.Infof("Attempting to remove peer %s (attempt %d/%d)", nodeID, i+1, maxRetries)
		future := m.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
		err := future.Error()
		if err == nil {
			klog.Infof("Successfully removed peer %s", nodeID)
			return nil
		}

		if i < maxRetries-1 {
			klog.Warningf("Failed to remove peer %s, will retry in %v: %v", nodeID, retryDelay, err)
			time.Sleep(retryDelay)
			retryDelay *= 2 // 指数退避
		}
	}

	return fmt.Errorf("failed to remove peer %s after %d attempts", nodeID, maxRetries)
}

// Shutdown 优雅关闭Raft实例
func (m *LeaderElectionManager) Shutdown(ctx context.Context) error {
	if m.raft == nil {
		return nil
	}

	klog.Infof("Shutting down leader election manager for node %s", m.config.NodeID)

	// 创建一个完成通道
	shutdownDone := make(chan struct{})

	// 在goroutine中执行shutdown以支持超时
	go func() {
		defer close(shutdownDone)
		// 关闭Raft实例
		m.raft.Shutdown()
		klog.Infof("Raft instance shut down for node %s", m.config.NodeID)
	}()

	// 等待shutdown完成或上下文超时
	select {
	case <-shutdownDone:
		// 成功关闭
		m.raft = nil
		return nil
	case <-ctx.Done():
		// 超时
		return fmt.Errorf("shutdown timed out: %v", ctx.Err())
	}
}

// GetRaftStatus 获取Raft当前状态
func (m *LeaderElectionManager) GetRaftStatus() raft.RaftState {
	if m.raft == nil {
		return raft.Shutdown
	}
	return m.raft.State()
}

// GetConfiguration 获取当前集群配置
func (m *LeaderElectionManager) GetConfiguration() (raft.Configuration, error) {
	if m.raft == nil {
		return raft.Configuration{}, fmt.Errorf("raft instance not initialized")
	}

	future := m.raft.GetConfiguration()
	if err := future.Error(); err != nil {
		return raft.Configuration{}, fmt.Errorf("failed to get configuration: %v", err)
	}

	return future.Configuration(), nil
}
