package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

type TaskConsumer struct {
	node           *DistributedNode
	workerCount    int
	taskQueue      chan Task
	wg             sync.WaitGroup
	mu             sync.Mutex
	totalTasks     int64
	completedTasks int64
	failedTasks    int64
	processingTime int64
}

type TaskResult struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewTaskConsumer(node *DistributedNode) *TaskConsumer {
	return &TaskConsumer{
		node:        node,
		workerCount: DefaultConcurrentWorkers,
		taskQueue:   make(chan Task, DefaultTaskQueueSize),
		mu:          sync.Mutex{},
	}
}

func (c *TaskConsumer) Start() {
	log.Printf("Consumer starting with %d workers on node %s",
		c.workerCount, c.node.config.NodeIP)

	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	// 启动任务拉取循环，使用WaitGroup跟踪
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.fetchTasks()
	}()
}

func (c *TaskConsumer) Stop() {
	close(c.taskQueue)
	c.wg.Wait()
}

func (c *TaskConsumer) fetchTasks() {
	ticker := time.NewTicker(time.Duration(DefaultTaskPullInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.pullTasksFromGlobalQueue(); err != nil {
				log.Printf("Failed to pull assigned tasks: %v", err)
			}
		case <-c.node.ctx.Done():

			return
		}
	}
}

func (c *TaskConsumer) pullTasksFromGlobalQueue() error {
	taskQueue := GetTaskQueue()
	select {
	case task, ok := <-taskQueue:
		if ok {
			// 有任务时仍然记录日志
			log.Printf("Pulled task %s from global queue", task.ID)
			c.taskQueue <- task
		}
	default:
		// 无任务时完全不打印日志，减少日志量
	}
	return nil
}

func (c *TaskConsumer) worker(id int) {
	defer c.wg.Done()

	for task := range c.taskQueue {
		log.Printf("Worker-%d processing task %s", id, task.ID)

		startTime := time.Now()
		result := c.processTaskWithTimeout(task)
		processTime := time.Since(startTime).Milliseconds()
		c.updateTaskStatus(task, result)

		c.mu.Lock()
		c.totalTasks++
		c.processingTime += processTime
		if result.Success {
			c.completedTasks++
		} else {
			c.failedTasks++
		}
		c.mu.Unlock()
	}
}

// processTaskWithTimeout 带超时的任务处理
func (c *TaskConsumer) processTaskWithTimeout(task Task) *TaskResult {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(DefaultWorkerTimeout)*time.Second)
	defer cancel()

	resultChan := make(chan *TaskResult, 1)
	go func() {
		resultChan <- c.processTask(task)
	}()

	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		log.Printf("Task %s processing timed out after %d seconds", task.ID, DefaultWorkerTimeout)
		return &TaskResult{
			Success: false,
			Error:   "Task processing timed out",
		}
	}
}

// updateTaskStatus 更新任务最终状态
func (c *TaskConsumer) updateTaskStatus(task Task, result *TaskResult) {
	status := service.PipelineJobStatusSuccess
	if !result.Success {
		status = service.PipelineJobStatusFailed
	}

	dao := orm.NewOrm()
	_, err := dao.Raw(`UPDATE pipeline_stage_jobs SET status = ? WHERE id = ?`, status, task.ID).Exec()
	if err != nil {
		log.Printf("Failed to update status for task %s: %v", task.ID, err)
	}

	// 同步更新Raft状态
	if c.node.IsLeader() {
		taskState := Task{
			Type:   task.Type,
			Status: status,
		}
		data, _ := json.Marshal(taskState)
		cmd := Command{Type: "task", Data: data}
		cmdBytes, _ := json.Marshal(cmd)
		c.node.raft.Apply(cmdBytes, DefaultTimeCost)
	}
}

// processTask 实际任务处理逻辑
func (c *TaskConsumer) processTask(task Task) *TaskResult {
	// 根据task.Type路由到不同处理器
	// build_job/checkpoint/deploy_env/test_job
	switch task.Type {
	case "build_job":
		return c.build_job(task)
	case "checkpoint":
		return c.checkpoint(task)
	case "deploy_env":
		return c.deploy_env(task)
	case "test_job":
		return c.test_job(task)
	default:
		return &TaskResult{Success: false, Error: "unknown task type"}
	}
}

func (c *TaskConsumer) build_job(task Task) *TaskResult {
	var buildJobData service.PipelineStageJobs

	if err := json.Unmarshal(task.Payload, &buildJobData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}
	// 任务处理
	// TODO()
	fmt.Printf("build_job command: %v", buildJobData.Parameters)

	// 更新数据库状态为已完成
	dao := orm.NewOrm()
	_, err := dao.Raw(`UPDATE pipeline_stage_jobs SET status = ? WHERE id = ?`, service.PipelineJobStatusSuccess, buildJobData.Id).Exec()
	if err != nil {
		log.Printf("Failed to update status for build_job task %s: %v", buildJobData.Id, err)
	}

	return &TaskResult{Success: true, Data: map[string]string{"status": "success"}}
}

func (c *TaskConsumer) checkpoint(task Task) *TaskResult {
	var checkPointData service.PipelineStageJobs

	if err := json.Unmarshal([]byte(task.Payload), &checkPointData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}

	// 任务处理
	// TODO()
	fmt.Printf("checkpoint command: %v", checkPointData.Parameters)

	// 更新数据库状态为已完成
	dao := orm.NewOrm()
	_, err := dao.Raw(`UPDATE pipeline_stage_jobs SET status = ? WHERE id = ?`, service.PipelineJobStatusSuccess, checkPointData.Id).Exec()
	if err != nil {
		log.Printf("Failed to update status for checkpoint task %s: %v", checkPointData.Id, err)
	}

	return &TaskResult{Success: true, Data: map[string]string{"status": "success"}}
}

func (c *TaskConsumer) deploy_env(task Task) *TaskResult {
	var deployEnvData service.PipelineStageJobs

	if err := json.Unmarshal([]byte(task.Payload), &deployEnvData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}
	// 任务处理
	// TODO()
	fmt.Printf("deploy_env command: %v", deployEnvData.Parameters)

	// 更新数据库状态为已完成
	dao := orm.NewOrm()
	_, err := dao.Raw(`UPDATE pipeline_stage_jobs SET status = ? WHERE id = ?`, service.PipelineJobStatusSuccess, deployEnvData.Id).Exec()
	if err != nil {
		log.Printf("Failed to update status for deploy_env task %s: %v", deployEnvData.Id, err)
	}

	return &TaskResult{Success: true, Data: map[string]string{"status": "success"}}
}

func (c *TaskConsumer) test_job(task Task) *TaskResult {
	var testJobData service.PipelineStageJobs

	if err := json.Unmarshal([]byte(task.Payload), &testJobData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}
	// 任务处理
	// TODO()
	fmt.Printf("test_job command: %v", testJobData.Parameters)

	// 更新数据库状态为已完成
	dao := orm.NewOrm()
	_, err := dao.Raw(`UPDATE pipeline_stage_jobs SET status = ? WHERE id = ?`, service.PipelineJobStatusSuccess, testJobData.Id).Exec()
	if err != nil {
		log.Printf("Failed to update status for test_job task %s: %v", testJobData.Id, err)
	}

	return &TaskResult{Success: true, Data: map[string]string{"status": "success"}}
}
