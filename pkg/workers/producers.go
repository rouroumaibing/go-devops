package workers

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

// TaskProducer 任务生产者
type TaskProducer struct {
	node          *DistributedNode
	producerCount int
	wg            sync.WaitGroup
}

func NewTaskProducer(node *DistributedNode) *TaskProducer {
	return &TaskProducer{
		node:          node,
		producerCount: DefaultProducerCount,
	}
}

// controller check if node is leader from db.
func IsLeader(nodeIP string) bool {
	dao := orm.NewOrm()
	var count int64
	err := dao.Raw(`SELECT COUNT(*) FROM leader_election WHERE leader_ip = ?`, nodeIP).QueryRow(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func (p *TaskProducer) Start() {
	InitTaskQueue(TaskQueueCHSize)
	InitClusterList()

	go p.RunProducer()

}

func (p *TaskProducer) RunProducer() {
	log.Printf("Producer started on leader %s", p.node.config.NodeIP)

	for producer := 0; producer < p.producerCount; producer++ {
		p.wg.Add(1)
		go func(producerID int) {
			defer p.wg.Done()
			log.Printf("Producer goroutine %d started on node %s", producerID, p.node.config.NodeIP)

			// 将ticker移到循环外部，避免每次循环都创建新的ticker
			ticker := time.NewTicker(time.Duration(DefaultTaskProduceInterval) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					isLeader := p.node.IsLeader()
					if isLeader {
						if err := p.produceTasks(); err != nil {
							log.Printf("Failed to produce tasks: %v", err)
						}
					}
				case <-p.node.ctx.Done():
					log.Printf("Producer %d shutting down", producerID)
					return
				}
			}
		}(producer)
	}
}

func (p *TaskProducer) produceTasks() error {

	// 1. 查询pending状态任务
	// 从表PipelineStageJobs查询status为wait的任务所有字段
	dao := orm.NewOrm()
	var jobs []service.PipelineStageJobs
	_, err := dao.Raw(`SELECT * FROM pipeline_stage_jobs WHERE status = 'wait'`).QueryRows(&jobs)
	if err != nil {
		log.Printf("Error querying wait tasks: %v", err)
		return err
	}
	if jobs == nil {
		return nil
	}

	ch := GetTaskQueue()

	// 2. 分配任务到全局队列
	for _, job := range jobs {
		log.Printf("Processing job: %s, Name: %s", job.Id, job.Name)

		// 3. 更新任务状态为processing
		// 先更新状态再分配，避免重复处理
		p.updateTaskStatus(job.Id, "processing")

		// 4. 分配任务
		if err := p.assignTask(job.Name, job.Id, "processing"); err != nil {
			log.Printf("Failed to assign task %s: %v", job.Id, err)
			// 即使Raft分配失败，也将任务放入队列，确保任务能够被处理
		}

		jobJSON, err := json.Marshal(job)
		if err != nil {
			log.Printf("Failed to marshal job %s: %v", job.Id, err)
			continue
		}

		// 5. 将任务放入到全局队列中
		select {
		case ch <- Task{
			Type:    job.Name,
			ID:      job.Id,
			Status:  "processing",
			Payload: jobJSON,
		}:
			log.Printf("Successfully added task %s to global queue", job.Id)
		default:
			log.Printf("Failed to add task %s to queue: queue is full", job.Id)
		}
	}

	return nil
}

func (p *TaskProducer) assignTask(taskType string, taskID string, status string) error {
	assignment := Task{
		Type:   taskType,
		ID:     taskID,
		Status: status,
	}

	data, _ := json.Marshal(assignment)
	cmd := Command{
		Type: "task",
		Data: data,
	}

	cmdData, _ := json.Marshal(cmd)

	// 通过Raft提交，确保所有节点对任务分配达成一致
	future := p.node.raft.Apply(cmdData, DefaultTimeCost)
	return future.Error()
}

// updateTaskStatus 更新数据库中的任务状态
func (p *TaskProducer) updateTaskStatus(taskID string, status string) {
	dao := orm.NewOrm()
	sql := `UPDATE pipeline_stage_jobs SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	result, err := dao.Raw(sql, status, taskID).Exec()
	if err != nil {
		log.Printf("Failed to update task %s status to %s: %v", taskID, status, err)
		return
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to get affected rows for task %s: %v", taskID, err)
	} else if affected == 0 {
		log.Printf("Warning: No rows affected when updating task %s status", taskID)
	} else {
		log.Printf("Successfully updated task %s status to %s", taskID, status)
	}
}

func (p *TaskProducer) Stop() {
	log.Printf("Producer stopped on leader %s", p.node.config.NodeIP)
	p.wg.Wait()
}
