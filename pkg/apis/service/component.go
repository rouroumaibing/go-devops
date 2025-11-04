package service

import "time"

// Component 组件信息表结构体
type Component struct {
	Id          string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name        string    `json:"name" orm:"column(name);size(100);null(false);description(组件名称)"`
	ServiceTree string    `json:"service_tree" orm:"column(service_tree);size(100);null;description(服务树路径)"`
	Owner       string    `json:"owner" orm:"column(owner);size(50);null(false);description(责任人)"`
	Description string    `json:"description" orm:"column(description);type(text);null;description(组件描述)"`
	RepoUrl     string    `json:"repo_url" orm:"column(repo_url);size(255);null(false);description(代码库地址)"`
	RepoBranch  string    `json:"repo_branch" orm:"column(repo_branch);size(50);null(false);description(代码分支)"`
	CreatedAt   time.Time `json:"created_at" orm:"column(created_at);type(datetime);auto_now_add;description(创建时间)"`
	UpdatedAt   time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);auto_now;description(更新时间)"`
}

type ComponentParameter struct {
	Id            string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	ComponentId   string    `json:"component_id" orm:"column(component_id);size(36);null(false);description(组件ID)"`
	EnvironmentId string    `json:"environment_id" orm:"column(environment_id);size(50);null(false);description(环境)"`
	ParamYaml     string    `json:"param_yaml" orm:"column(param_yaml);type(text);null;description(参数yaml)"`
	CreatedAt     time.Time `json:"created_at" orm:"column(created_at);type(datetime);auto_now_add;description(创建时间)"`
	UpdatedAt     time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);auto_now;description(更新时间)"`
}
