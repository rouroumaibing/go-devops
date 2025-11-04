package worker

import (
	"fmt"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
	"github.com/rouroumaibing/go-devops/pkg/workers/common"
)

var (
	globalWorker *Worker
	workerOnce   sync.Once
)

type Worker struct {
	WorkerIP                string
	WorkerTimeout           int
	WorkerKeepAliveInterval int
	stopCh                  chan struct{}

	// 任务队列
	taskQueue chan *service.PipelineStageJobs
}

// 获取全局Worker实例
func GetWorker() *Worker {
	workerOnce.Do(func() {
		globalWorker = &Worker{
			stopCh:                  make(chan struct{}),
			WorkerTimeout:           common.DefaultWorkerTimeout,
			WorkerKeepAliveInterval: common.DefaultWorkerKeepAliveInterval,
			taskQueue:               make(chan *service.PipelineStageJobs, 1000),
		}
	})
	return globalWorker
}

// 初始化Worker
func initWorker() {
	worker := GetWorker()
	ip, err := system.GetLocalIP()
	if err != nil {
		fmt.Printf("获取本地IP失败: %v\n", err)
		return
	}
	worker.WorkerIP = ip

	// 启动多个工作协程处理任务
	for i := 0; i < common.DefaultWorkerCount; i++ {
		go worker.processTasks(i)
	}

	// 启动心跳
	go worker.keepAlive()
}

// 处理任务
func (w *Worker) processTasks(workerID int) {
	for {
		select {
		case <-w.stopCh:
			return
		case task := <-w.taskQueue:
			fmt.Printf("Worker-%d 开始处理任务: %s\n", workerID, task.Id)
			w.executeTask(task)
		}
	}
}

// 执行具体任务
func (w *Worker) executeTask(task *service.PipelineStageJobs) {
	// 这里是实际执行任务的逻辑
	// 模拟任务执行
	time.Sleep(5 * time.Second)

	// 更新任务状态
	o := orm.NewOrm()
	task.Status = service.PipelineJobStatusSuccess
	task.UpdatedAt = time.Now()
	_, err := o.Update(task)
	if err != nil {
		fmt.Printf("更新任务状态失败: %v\n", err)
	}

	fmt.Printf("任务执行完成: %s, 状态: %s\n", task.Id, task.Status)
}

// 心跳保持
func (w *Worker) keepAlive() {
	ticker := time.NewTicker(time.Duration(w.WorkerKeepAliveInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			// 向master报告worker的状态
			// 这里简化处理，只打印日志
			fmt.Printf("Worker %s 心跳\n", w.WorkerIP)
		}
	}
}

// 守护进程
func Deamon() {
	// 初始化Worker
	initWorker()

	// 模拟从数据库轮询任务
	go func() {
		for {
			select {
			case <-GetWorker().stopCh:
				return
			default:
				// 检查任务队列中是否有任务
				task := GetWorker().fetchTaskFromDB()
				if task != nil {
					GetWorker().taskQueue <- task
				} else {
					// 没有任务则循环等待10s，再检查一次，直到任务队列中有任务
					time.Sleep(10 * time.Second)
				}
			}
		}
	}()

	// 持续运行
	<-GetWorker().stopCh
}

// 从数据库获取任务
func (w *Worker) fetchTaskFromDB() *service.PipelineStageJobs {
	o := orm.NewOrm()
	task := &service.PipelineStageJobs{}

	// 开始事务
	err := o.Begin()
	if err != nil {
		fmt.Printf("开始事务失败: %v\n", err)
		return nil
	}
	defer func() {
		// 检查事务是否已提交，如果没有提交则回滚
		if err = o.Rollback(); err != nil {
			// 注意：这里只记录日志，不影响返回值，因为事务回滚失败通常是因为事务已经被提交或回滚
			fmt.Printf("事务回滚失败: %v\n", err)
		}
	}()

	// 查询并锁定任务
	err = o.QueryTable(new(service.PipelineStageJobs)).
		Filter("status", service.PipelineJobStatusRun).
		Limit(1).
		ForUpdate().
		One(task)

	if err != nil {
		// 如果没有任务，不会打印错误
		if err != orm.ErrNoRows {
			fmt.Printf("查询任务失败: %v\n", err)
		}
		return nil
	}

	// 提交事务
	if err := o.Commit(); err != nil {
		fmt.Printf("提交事务失败: %v\n", err)
		return nil
	}

	return task
}

// 停止Worker
func StopWorker() {
	worker := GetWorker()
	close(worker.stopCh)
}
