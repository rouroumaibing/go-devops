package workers

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
)

type Command struct {
	Type       string          `json:"type"` // task_update, leader_info, task_assign
	Data       json.RawMessage `json:"data,omitempty"`
	LeaderID   string          `json:"leader_id,omitempty"`
	LeaderAddr string          `json:"leader_addr,omitempty"`
	Term       uint64          `json:"term,omitempty"`
}

type TaskAssignment struct {
	TaskID   int64  `json:"task_id"`
	WorkerID string `json:"worker_id"`
}

type ApplyResult struct {
	Success bool
	Message string
}

// TaskState 任务状态
type TaskState struct {
	TaskID    int64  `json:"task_id"`
	Status    string `json:"status"`
	WorkerID  string `json:"worker_id"`
	Timestamp int64  `json:"timestamp"`
}

// ClusterState FSM状态机
type ClusterState struct {
	mu          sync.RWMutex
	tasks       map[int64]*TaskState
	currentTerm uint64
	leaderID    string
	leaderAddr  string
	clusterID   string
}

type ClusterSnapshot struct {
	Tasks       map[int64]*TaskState
	CurrentTerm uint64
	LeaderID    string
	LeaderAddr  string
}

func NewClusterState(clusterID string) *ClusterState {
	return &ClusterState{
		tasks:     make(map[int64]*TaskState),
		clusterID: clusterID,
	}
}

func (cs *ClusterState) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	switch cmd.Type {
	case "task_update":
		// 更新任务状态
		var task TaskState
		if err := json.Unmarshal(cmd.Data, &task); err != nil {
			return fmt.Errorf("unmarshal task error: %v", err)
		}
		cs.tasks[task.TaskID] = &task
		return &ApplyResult{Success: true}

	case "leader_info":
		// 更新Leader信息
		cs.leaderID = cmd.LeaderID
		cs.leaderAddr = cmd.LeaderAddr
		cs.currentTerm = cmd.Term
		return &ApplyResult{Success: true}

	case "task_assign":
		// 分配任务给worker
		var assignment TaskAssignment
		if err := json.Unmarshal(cmd.Data, &assignment); err != nil {
			return fmt.Errorf("unmarshal assignment error: %v", err)
		}
		if task, ok := cs.tasks[assignment.TaskID]; ok {
			task.WorkerID = assignment.WorkerID
			task.Status = "processing"
		}
		return &ApplyResult{Success: true}
	}
	return fmt.Errorf("unknown command type")
}

func (cs *ClusterState) Snapshot() (raft.FSMSnapshot, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	snap := &ClusterSnapshot{
		Tasks:       cs.tasks,
		CurrentTerm: cs.currentTerm,
		LeaderID:    cs.leaderID,
		LeaderAddr:  cs.leaderAddr,
	}
	return snap, nil
}

func (cs *ClusterState) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var snap ClusterSnapshot
	if err := json.NewDecoder(rc).Decode(&snap); err != nil {
		return err
	}

	cs.mu.Lock()
	cs.tasks = snap.Tasks
	cs.currentTerm = snap.CurrentTerm
	cs.leaderID = snap.LeaderID
	cs.leaderAddr = snap.LeaderAddr
	cs.mu.Unlock()
	return nil
}

// Persist 实现raft.FSMSnapshot接口的Persist方法
// 将快照数据写入到指定的接收器
func (cs *ClusterSnapshot) Persist(sink raft.SnapshotSink) error {
	// 将快照数据编码为JSON
	if err := json.NewEncoder(sink).Encode(cs); err != nil {
		if err = sink.Cancel(); err != nil {
			return err
		}
		return err
	}
	// 成功完成快照保存
	return sink.Close()
}

// Release 实现raft.FSMSnapshot接口的Release方法
// 用于释放资源
func (cs *ClusterSnapshot) Release() {
	// 在这里可以添加需要释放的资源
}
