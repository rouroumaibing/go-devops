package workers

import (
	"time"

	"github.com/astaxie/beego/config/env"
)

const (
	DefaultTimeCost                = 5 * time.Second
	DefaultConcurrentWorkers       = 10
	DefaultGlobalMaxQPS            = 50
	DefaultGlobalMaxBurst          = 50
	DefaultWorkerCount             = 5
	DefaultMasterTimeout           = 10
	DefaultMasterElectionInterval  = 2 // 主节点选举间隔时间
	DefaultWorkerTimeout           = 10
	DefaultWorkerKeepAliveInterval = 2  // 从节点心跳间隔时间
	DefaultTaskPullInterval        = 3  // 任务拉取间隔时间（秒）
	DefaultTaskProduceInterval     = 2  // 任务生产间隔时间（秒）
	DefaultDBConnMaxLifetime       = 5  // 数据库连接最大生命周期（分钟）
	DefaultRaftSnapshotInterval    = 30 // Raft快照间隔（秒）
	DefaultRaftTransportTimeout    = 10 // Raft传输超时（秒）
)

type ControllerOption struct {
	TimeCost          time.Duration // 任务处理时间超时时间
	ConcurrentWorkers int           // 任务并发的worker数量
	GlobalMaxQPS      int           // 全局最大处理任务数量
	GlobalMaxBurst    int           // 全局最大任务突发数量
}

func GetClusterID() string {
	return env.Get("CLUSTERID", "default-cluster")
}
