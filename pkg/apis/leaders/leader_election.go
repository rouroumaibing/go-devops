package leaders

import (
	"time"
)

type LeaderElection struct {
	Id            string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	RaftClusterID string    `json:"raft_clusterid" orm:"column(raft_clusterid);size(255);null;description(raft集群ID)"`
	LeaderID      string    `json:"leader_id" orm:"column(leader_id);size(255);null;description(主节点, 格式为:集群ID@节点IP)"`
	LeaderIP      string    `json:"leader_ip" orm:"column(leader_ip);size(255);null;description(主节点IP)"`
	LeaderAddr    string    `json:"leader_addr" orm:"column(leader_addr);size(255);null;description(节点+端口)"`
	RaftTerm      string    `json:"raft_term" orm:"column(raft_term);size(255);null;description(raft term)"`
	LastHeartbeat time.Time `json:"last_heartbeat" orm:"column(last_heartbeat);type(datetime);null;description(最后心跳时间)"`
	CreatedAt     time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt     time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}
