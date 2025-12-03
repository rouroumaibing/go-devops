package workers

import (
	"encoding/json"
	"sync"
	"time"
)

const (
	DefaultTimeCost                    = 5 * time.Second // 任务处理时间超时时间
	DefaultConcurrentWorkers           = 10              // 任务并发的worker数量
	DefaultGlobalMaxQPS                = 50              // 全局最大处理任务数量
	DefaultGlobalMaxBurst              = 50              // 全局最大任务突发数量
	DefaultProducerCount               = 3               // 任务生产者数量
	DefaultWorkerCount                 = 5               // 任务并发的worker数量
	DefaultTaskQueueSize               = 100
	DefaultMasterTimeout               = 10              // 主节点超时时间
	DefaultMasterElectionInterval      = 2               // 主节点选举间隔时间
	DefaultWorkerTimeout               = 10              // 从节点超时时间
	DefaultWorkerKeepAliveInterval     = 2               // 从节点心跳间隔时间
	DefaultTaskPullInterval            = 3               // 任务拉取间隔时间（秒）
	DefaultTaskProduceInterval         = 2               // 任务生产间隔时间（秒）
	DefaultDBConnMaxLifetime           = 5               // 数据库连接最大生命周期（分钟）
	DefaultRaftSnapshotInterval        = 30              // Raft快照间隔（秒）
	DefaultRaftTransportTimeout        = 10              // Raft传输超时（秒）
	DefaultRaftHeartbeatInterval       = 1               // Raft心跳间隔（秒）
	DefaultRaftElectionTimeout         = 3               // Raft选举超时（秒）
	DefaultRaftCommitTimeout           = 50              // Raft提交超时（毫秒）
	DefaultRaftNodeHealthCheckInterval = 5               // Raft节点健康检查间隔（秒）
	DefaultLeaderElectionUpdateTimeout = 5 * time.Minute // 更新选主记录超时时间
)
const (
	DefaultRaftPort = "7001"
	DefaultDataDir  = "./data"
)

type Command struct {
	Type     string          `json:"type"` // leader_update, task
	Data     json.RawMessage `json:"data,omitempty"`
	LeaderID string          `json:"leader_id,omitempty"`
	LeaderIP string          `json:"leader_ip,omitempty"`
	Term     uint64          `json:"term,omitempty"`
}

type ApplyResult struct {
	Success bool
	Message string
}

type Task struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Status  string `json:"status"`
	Payload []byte `json:"payload"`
}

var (
	JoinClusterList       []*NodeConfig // 待加入集群节点列表
	JoinClusterListSize   = 10
	RemoveClusterList     []*NodeConfig // 待移除集群节点列表
	RemoveClusterListSize = 10
)

var (
	TaskQueueCH     chan Task // 全局任务队列管道
	TaskQueueCHOnce sync.Once // 初始化任务队列的Once对象
	TaskQueueCHSize = 100     // 任务队列缓冲区大小

)

func InitClusterList() {
	JoinClusterList = make([]*NodeConfig, JoinClusterListSize)
	RemoveClusterList = make([]*NodeConfig, RemoveClusterListSize)
}

func InitTaskQueue(bufferSize int) {
	TaskQueueCHOnce.Do(func() {
		TaskQueueCH = make(chan Task, bufferSize)
	})
}

func GetTaskQueue() chan Task {
	if TaskQueueCH == nil {
		InitTaskQueue(TaskQueueCHSize)
	}
	return TaskQueueCH
}
