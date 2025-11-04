package leaders

import "time"

type LeaderElection struct {
	Id        string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Kind      string    `json:"kind" orm:"column(kind);size(255);null;description(类型,例如master/worker)"`
	ClusterID string    `json:"clusterID" orm:"column(cluster_id);size(255);null;description(集群ID)"`
	LeaderIP  string    `json:"leaderIP" orm:"column(leader_ip);size(255);null;description(主节点IP)"`
	CreatedAt time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}
