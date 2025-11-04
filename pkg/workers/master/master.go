package master

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
	"github.com/rouroumaibing/go-devops/pkg/workers/common"
)

const (
	defaultClusterID = "default-cluster"
)

var (
	globalMaster *Master
	once         sync.Once
)

type Master struct {
	sync.RWMutex
	IP                     string
	MasterTimeout          int
	MasterElectionInterval int

	ThisClusterID   string
	LeaderClusterID string

	stopCh    chan struct{}
	MasterIP  string
	taskQueue chan *service.PipelineStageJobs
}

func GetMaster() *Master {
	once.Do(func() {
		globalMaster = &Master{
			stopCh:                 make(chan struct{}),
			ThisClusterID:          defaultClusterID,
			LeaderClusterID:        defaultClusterID,
			MasterTimeout:          common.DefaultMasterTimeout,
			MasterElectionInterval: common.DefaultMasterElectionInterval,
			taskQueue:              make(chan *service.PipelineStageJobs, 1000),
		}
	})
	return globalMaster
}

func initMaster() {
	master := GetMaster()
	ip, err := system.GetLocalIP()
	if err != nil {
		fmt.Printf("获取本地IP失败: %v\n", err)
		return
	}
	master.IP = ip
	master.MasterIP = ip

	// 启动领导者选举
	go func() {
		for {
			if err := models.StartElection(); err != nil {
				fmt.Printf("领导者选举失败: %v\n", err)
			}
			time.Sleep(time.Duration(master.MasterElectionInterval) * time.Second)
		}
	}()

	// 启动心跳
	models.HeartBeat()

	// 启动任务处理
	go master.ProcessTasks()
}

// 处理任务队列中的任务
func (m *Master) ProcessTasks() {
	for {
		select {
		case <-m.stopCh:
			return
		case task := <-m.taskQueue:
			// 实际环境中，这里应该将任务分发到worker节点
			// 这里简化处理，直接更新任务状态
			o := orm.NewOrm()
			task.Status = service.PipelineJobStatusRun
			task.UpdatedAt = time.Now()
			_, err := o.Update(task)
			if err != nil {
				fmt.Printf("更新任务状态失败: %v\n", err)
			}
			fmt.Printf("处理任务: %s, 状态更新为: %s\n", task.Id, task.Status)
		}
	}
}

// 启动的时候检查PipelineStageGroupJobs、PipelineStageJobs的任务是否已经完成，没有完成加载到任务队列中
func (m *Master) RecoverJobs() {
	var err error
	// 查询未完成的任务
	var stageJobs []service.PipelineStageJobs
	o := orm.NewOrm()

	// 使用Raw SQL执行OR条件查询
	_, err = o.Raw("SELECT * FROM pipeline_stage_jobs WHERE status = ? OR status = ?",
		service.PipelineJobStatusInit, "processing").QueryRows(&stageJobs)

	if err != nil {
		fmt.Printf("查询未完成任务失败: %v\n", err)
		return
	}

	// 将未完成的任务加入队列
	master := GetMaster()
	for i := range stageJobs {
		master.taskQueue <- &stageJobs[i]
	}

	fmt.Printf("恢复了 %d 个未完成的任务\n", len(stageJobs))
}

func Deamon() {
	// 初始化Master
	initMaster()
	// 启动的时候检查PipelineStageGroupJobs、PipelineStageJobs的任务是否已经完成，没有完成加载到任务队列中
	GetMaster().RecoverJobs()

	// 持续运行
	<-GetMaster().stopCh
}

func StopMaster() {
	close(GetMaster().stopCh)
	models.StopHeartBeat()
}
