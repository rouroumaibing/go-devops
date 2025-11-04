package common

import (
	"time"

	"github.com/astaxie/beego/config/env"
)

const (
	MasterKind = "master"
	WorkerKind = "worker"
)

const (
	DefaultTimeCost                = 5 * time.Second
	DefaultConcurrentWorkers       = 10
	DefaultGolbalMaxQPS            = 50
	DefaultGolbalMaxBurst          = 50
	DefaultWorkerCount             = 5
	DefaultMasterTimeout           = 10
	DefaultMasterElectionInterval  = 2 // 主节点选举间隔时间
	DefaultWorkerTimeout           = 10
	DefaultWorkerKeepAliveInterval = 2 // 从节点心跳间隔时间
)

type ControllerOption struct {
	TimeCost          time.Duration // 任务处理时间超时时间
	ConcurrentWorkers int           // 任务并发的worker数量
	GolbalMaxQPS      int           // 全局最大处理任务数量
	GolbalMaxBurst    int           // 全局最大任务突发数量
}

func GetClusterID() string {
	return env.Get("CLUSTERID", "default-cluster")
}
