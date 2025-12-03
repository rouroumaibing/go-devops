package workers

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/hashicorp/raft"
	"k8s.io/klog/v2"
)

type ClusterSnapshot struct {
	Tasks       map[string]*Task
	CurrentTerm uint64
	LeaderID    string
	LeaderIP    string
}

func (cs *ClusterSnapshot) Persist(sink raft.SnapshotSink) error {
	defer func() {
		if r := recover(); r != nil {
			_ = sink.Cancel()
			klog.Errorf("Snapshot persist panic: %v", r)
		}
	}()

	if err := json.NewEncoder(sink).Encode(cs); err != nil {
		_ = sink.Cancel()
		return fmt.Errorf("failed to encode snapshot: %w", err)
	}

	return sink.Close()
}

func (cs *ClusterSnapshot) Release() {

}

type ClusterFSM struct {
	mu          sync.RWMutex
	tasks       map[string]*Task
	currentTerm uint64
	leaderID    string
	leaderIP    string
}

func NewClusterFSM(raftClusterID string) *ClusterFSM {
	return &ClusterFSM{
		tasks:    make(map[string]*Task),
		leaderID: raftClusterID,
	}
}

func (cfsm *ClusterFSM) Apply(log *raft.Log) interface{} {
	var cmd Command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return fmt.Errorf("unmarshal error: %v", err)
	}

	cfsm.mu.Lock()
	defer cfsm.mu.Unlock()

	switch cmd.Type {
	case "leader_update":
		// 更新Leader信息 - 保存完整的组合格式leaderID
		cfsm.leaderID = cmd.LeaderID
		cfsm.leaderIP = cmd.LeaderIP
		cfsm.currentTerm = cmd.Term
		return &ApplyResult{Success: true}

	case "task":
		var assignment Task
		if err := json.Unmarshal(cmd.Data, &assignment); err != nil {
			return fmt.Errorf("unmarshal assignment error: %v", err)
		}
		if task, ok := cfsm.tasks[assignment.ID]; ok {
			task.Type = assignment.Type
			task.ID = assignment.ID
			task.Status = "processing"
		}
		return &ApplyResult{Success: true}
	}
	return fmt.Errorf("unknown command type")
}

func (cfsm *ClusterFSM) Snapshot() (raft.FSMSnapshot, error) {
	cfsm.mu.RLock()
	defer cfsm.mu.RUnlock()

	snap := &ClusterSnapshot{
		Tasks:       cfsm.tasks,
		CurrentTerm: cfsm.currentTerm,
		LeaderID:    cfsm.leaderID,
		LeaderIP:    cfsm.leaderIP,
	}
	return snap, nil
}

func (cfsm *ClusterFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var snap ClusterSnapshot
	if err := json.NewDecoder(rc).Decode(&snap); err != nil {
		return err
	}

	cfsm.mu.Lock()
	cfsm.tasks = snap.Tasks
	cfsm.currentTerm = snap.CurrentTerm
	cfsm.leaderID = snap.LeaderID
	cfsm.leaderIP = snap.LeaderIP
	cfsm.mu.Unlock()
	return nil
}
