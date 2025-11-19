package workers

import (
	"encoding/json"
	"log"
	"math/rand"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type Task struct {
	ID      int64           `json:"id"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Retry   int             `json:"retry"`
}

// TaskProducer 任务生产者
type TaskProducer struct {
	node       *DistributedNode
	workerPool []string
	batchSize  int
	mu         sync.RWMutex
	limiter    *rate.Limiter
}

func NewTaskProducer(node *DistributedNode) *TaskProducer {
	return &TaskProducer{
		node:       node,
		workerPool: []string{node.config.NodeID}, // 初始只有自己
		batchSize:  10,
		// 使用默认的全局QPS和Burst配置初始化限流器
		limiter: rate.NewLimiter(rate.Limit(DefaultGlobalMaxQPS), DefaultGlobalMaxBurst),
	}
}

// runProducer 主生产循环
func (p *TaskProducer) RunProducer() {
	log.Printf("Producer started on leader %s", p.node.config.NodeID)

	ticker := time.NewTicker(time.Duration(DefaultTaskProduceInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !p.node.IsLeader() {
				log.Printf("No longer leader, stopping producer")
				return
			}
			if err := p.produceTasks(); err != nil {
				log.Printf("Failed to produce tasks: %v", err)
			}
		case <-p.node.ctx.Done():
			return
		}
	}
}

// produceTasks 从MySQL拉取任务并分配
func (p *TaskProducer) produceTasks() error {
	// 应用速率限制
	if err := p.limiter.Wait(p.node.ctx); err != nil {
		log.Printf("Rate limiter error: %v", err)
		return err
	}

	// 1. 查询pending状态任务
	query := `SELECT id, task_type, payload, retry_count 
              FROM tasks 
              WHERE cluster_id = ? AND status = 'pending' 
              LIMIT ?`

	rows, err := p.node.mysqlDB.Query(query, p.node.config.ClusterID, p.batchSize)
	if err != nil {
		return err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Type, &task.Payload, &task.Retry)
		if err != nil {
			log.Printf("Failed to scan task: %v", err)
			continue
		}
		tasks = append(tasks, task)
	}
	if err := rows.Err(); err != nil {
		log.Printf("Error iterating tasks: %v", err)
	}

	// 2. 分配任务给worker
	for _, task := range tasks {
		workerID := p.selectWorker() // 负载均衡策略
		if err := p.assignTask(task, workerID); err != nil {
			log.Printf("Failed to assign task %d: %v", task.ID, err)
			continue
		}

		// 3. 更新任务状态为processing
		p.updateTaskStatus(task.ID, "processing", workerID)
	}

	return nil
}

// selectWorker 选择worker节点（简单轮询或一致性哈希）
func (p *TaskProducer) selectWorker() string {
	// 简单随机策略
	// 生产环境建议使用一致性哈希：taskID -> workerID
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.workerPool[rand.Intn(len(p.workerPool))]
}

// assignTask 通过Raft提交任务分配命令
func (p *TaskProducer) assignTask(task Task, workerID string) error {
	assignment := TaskAssignment{
		TaskID:   task.ID,
		WorkerID: workerID,
	}

	data, _ := json.Marshal(assignment)
	cmd := Command{
		Type: "task_assign",
		Data: data,
	}

	cmdData, _ := json.Marshal(cmd)

	// 通过Raft提交，确保所有节点对任务分配达成一致
	future := p.node.raft.Apply(cmdData, DefaultTimeCost)
	return future.Error()
}

// updateTaskStatus 更新MySQL中的任务状态
func (p *TaskProducer) updateTaskStatus(taskID int64, status string, workerID string) {
	sql := `UPDATE tasks SET status = ?, worker_id = ? WHERE id = ?`
	_, err := p.node.mysqlDB.Exec(sql, status, workerID, taskID)
	if err != nil {
		log.Printf("Failed to update task status: %v", err)
	}
}

// AddWorker 动态添加worker节点
func (p *TaskProducer) AddWorker(nodeID string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, id := range p.workerPool {
		if id == nodeID {
			return // 已存在
		}
	}
	p.workerPool = append(p.workerPool, nodeID)
	log.Printf("Worker %s added to pool", nodeID)
}
