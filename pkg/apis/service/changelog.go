package service

import "time"

// ChangeLog 变更记录表结构体
type ChangeLog struct {
	Id             string    `json:"id" orm:"pk;szie(36);column(id);type(string);unique;description(主键ID)"`
	ComponentId    string    `json:"component_id" orm:"column(component_id);null(false);description(关联组件ID)"`
	ChangeNumber   string    `json:"change_number" orm:"column(change_number);size(50);null(false);description(变更编号)"`
	Description    string    `json:"description" orm:"column(description);type(text);null;description(变更描述)"`
	Content        string    `json:"content" orm:"column(content);type(text);null;description(变更内容)"`
	Status         string    `json:"status" orm:"column(status);size(20);null(false);description(变更状态:pending,processing,success,failed)"`
	Creator        string    `json:"creator" orm:"column(creator);size(50);null(false);description(提交人)"`
	LastUpdateTime time.Time `json:"last_update_time" orm:"column(last_update_time);type(datetime);auto_now;description(最后更新时间)"`
}
