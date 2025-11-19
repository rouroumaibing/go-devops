package workers

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"
)

const DefaultTaskQueueSize = 100

// TaskConsumer 任务消费者
type TaskConsumer struct {
	node        *DistributedNode
	workerCount int
	taskQueue   chan Task
	wg          sync.WaitGroup
}

type TaskResult struct {
	Success bool        `json:"success"`
	Error   string      `json:"error,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// NewTaskConsumer 创建任务消费者，使用默认worker数量
func NewTaskConsumer(node *DistributedNode) *TaskConsumer {
	return NewTaskConsumerWithWorkerCount(node, DefaultWorkerCount)
}

// NewTaskConsumerWithWorkerCount 创建指定worker数量的任务消费者
func NewTaskConsumerWithWorkerCount(node *DistributedNode, workerCount int) *TaskConsumer {
	// 如果workerCount为0，使用默认值
	if workerCount <= 0 {
		workerCount = DefaultConcurrentWorkers
	}
	return &TaskConsumer{
		node:        node,
		workerCount: workerCount,
		taskQueue:   make(chan Task, DefaultTaskQueueSize),
	}
}

// NewTaskConsumerWithQueueSize 创建指定worker数量和队列大小的任务消费者
func NewTaskConsumerWithQueueSize(node *DistributedNode, workerCount int, queueSize int) *TaskConsumer {
	if workerCount <= 0 {
		workerCount = DefaultConcurrentWorkers
	}
	if queueSize <= 0 {
		queueSize = DefaultTaskQueueSize
	}
	return &TaskConsumer{
		node:        node,
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
	}
}

// NewTaskConsumerWithOptions 创建支持ControllerOption配置的任务消费者
func NewTaskConsumerWithOptions(node *DistributedNode, options *ControllerOption) *TaskConsumer {
	// 使用默认值作为备选
	workerCount := DefaultConcurrentWorkers
	queueSize := DefaultTaskQueueSize

	// 如果提供了选项，则使用选项中的值
	if options != nil {
		if options.ConcurrentWorkers > 0 {
			workerCount = options.ConcurrentWorkers
		}
		// 注意：ControllerOption中没有队列大小配置，仍使用默认值
	}

	return &TaskConsumer{
		node:        node,
		workerCount: workerCount,
		taskQueue:   make(chan Task, queueSize),
	}
}

// Start 启动消费者
func (c *TaskConsumer) Start() {
	log.Printf("Consumer starting with %d workers on node %s",
		c.workerCount, c.node.config.NodeID)

	// 启动worker协程
	for i := 0; i < c.workerCount; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}

	// 启动任务拉取循环
	go c.fetchTasks()
}

// Stop 停止消费者
func (c *TaskConsumer) Stop() {
	close(c.taskQueue)
	c.wg.Wait()
}

// fetchTasks 从MySQL拉取分配给本节点的任务
func (c *TaskConsumer) fetchTasks() {
	ticker := time.NewTicker(time.Duration(DefaultTaskPullInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.pullAssignedTasks(); err != nil {
				log.Printf("Failed to pull assigned tasks: %v", err)
			}
		case <-c.node.ctx.Done():
			return
		}
	}
}

// pullAssignedTasks 查询分配给本节点的pending任务
func (c *TaskConsumer) pullAssignedTasks() error {
	query := `SELECT id, task_type, payload, retry_count 
              FROM tasks 
              WHERE cluster_id = ? AND worker_id = ? AND status = 'pending' 
              LIMIT 20`

	rows, err := c.node.mysqlDB.Query(query, c.node.config.ClusterID, c.node.config.NodeID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Type, &task.Payload, &task.Retry); err != nil {
			log.Printf("Failed to scan task: %v", err)
			continue
		}
		c.taskQueue <- task // 放入队列
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating tasks: %v", err)
	}
	return nil
}

// worker 任务处理worker
func (c *TaskConsumer) worker(id int) {
	defer c.wg.Done()

	for task := range c.taskQueue {
		log.Printf("Worker-%d processing task %d", id, task.ID)

		// 1. 处理任务 - 添加超时控制
		result := c.processTaskWithTimeout(task)

		// 2. 记录处理结果
		c.recordResult(task, result)

		// 3. 更新状态到MySQL
		c.updateTaskStatus(task, result)
	}
}

// processTaskWithTimeout 带超时的任务处理
func (c *TaskConsumer) processTaskWithTimeout(task Task) *TaskResult {
	// 创建一个带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(DefaultWorkerTimeout)*time.Second)
	defer cancel()

	// 创建一个通道接收处理结果
	resultChan := make(chan *TaskResult, 1)

	go func() {
		resultChan <- c.processTask(task)
	}()

	// 等待处理结果或超时
	select {
	case result := <-resultChan:
		return result
	case <-ctx.Done():
		// 任务处理超时
		log.Printf("Task %d processing timed out after %d seconds", task.ID, DefaultWorkerTimeout)
		return &TaskResult{
			Success: false,
			Error:   "Task processing timed out",
		}
	}
}

// processTask 实际任务处理逻辑
func (c *TaskConsumer) processTask(task Task) *TaskResult {
	// 根据task.Type路由到不同处理器
	switch task.Type {
	case "email_send":
		return c.handleEmailTask(task)
	case "data_sync":
		return c.handleDataSyncTask(task)
	// ... 其他任务类型
	default:
		return &TaskResult{Success: false, Error: "unknown task type"}
	}
}

// handleEmailTask 示例：处理邮件发送任务
func (c *TaskConsumer) handleEmailTask(task Task) *TaskResult {
	var emailData struct {
		To      string `json:"to"`
		Subject string `json:"subject"`
		Body    string `json:"body"`
	}

	if err := json.Unmarshal(task.Payload, &emailData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}

	// 模拟处理
	time.Sleep(1 * time.Second)
	log.Printf("Email sent to %s", emailData.To)

	return &TaskResult{Success: true, Data: map[string]string{"status": "sent"}}
}

// handleDataSyncTask 示例：处理数据同步任务
func (c *TaskConsumer) handleDataSyncTask(task Task) *TaskResult {
	var syncData struct {
		Source      string `json:"source"`
		Destination string `json:"destination"`
		DataType    string `json:"data_type"`
		BatchSize   int    `json:"batch_size"`
	}

	if err := json.Unmarshal(task.Payload, &syncData); err != nil {
		return &TaskResult{Success: false, Error: err.Error()}
	}

	// 模拟处理
	time.Sleep(2 * time.Second) // 假设数据同步需要更长时间
	log.Printf("Data sync completed from %s to %s, data type: %s, batch size: %d",
		syncData.Source, syncData.Destination, syncData.DataType, syncData.BatchSize)

	return &TaskResult{Success: true, Data: map[string]interface{}{
		"status":       "completed",
		"source":       syncData.Source,
		"destination":  syncData.Destination,
		"records_sync": syncData.BatchSize, // 模拟同步的记录数
	}}
}

// recordResult 记录处理结果到MySQL（幂等性保障）
func (c *TaskConsumer) recordResult(task Task, result *TaskResult) {
	data, _ := json.Marshal(result)
	sql := `INSERT INTO task_processed (task_id, node_id, result) VALUES (?, ?, ?)`
	_, err := c.node.mysqlDB.Exec(sql, task.ID, c.node.config.NodeID, data)
	if err != nil {
		log.Printf("Failed to record result: %v", err)
	}
}

// updateTaskStatus 更新任务最终状态
func (c *TaskConsumer) updateTaskStatus(task Task, result *TaskResult) {
	status := "completed"
	if !result.Success {
		status = "failed"
		// 失败重试逻辑
		if task.Retry < 3 {
			status = "pending"
			if _, err := c.node.mysqlDB.Exec(`UPDATE tasks SET retry_count = retry_count + 1 WHERE id = ?`, task.ID); err != nil {
				log.Printf("Failed to update retry count for task %d: %v", task.ID, err)
			}
		}
	}

	if _, err := c.node.mysqlDB.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, status, task.ID); err != nil {
		log.Printf("Failed to update status for task %d: %v", task.ID, err)
	}

	// 同步更新Raft状态
	if c.node.IsLeader() {
		taskState := TaskState{
			TaskID:    task.ID,
			Status:    status,
			WorkerID:  c.node.config.NodeID,
			Timestamp: time.Now().Unix(),
		}
		data, _ := json.Marshal(taskState)
		cmd := Command{Type: "task_update", Data: data}
		cmdBytes, _ := json.Marshal(cmd)
		c.node.raft.Apply(cmdBytes, DefaultTimeCost)
	}
}
