package service

import (
	"time"

	"github.com/google/uuid"
)

// Environment 环境配置表结构体
type Environment struct {
	Id              string         `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name            string         `json:"name" orm:"column(name);size(50);null(false);description(环境名称)"`
	IsProd          bool           `json:"is_prod" orm:"column(is_prod);null(false);default(false);description(是否为生产环境)"`
	IsEnv           bool           `json:"is_env" orm:"column(is_env);null(false);default(false);description(是否为环境节点)"`
	EnvGroup        string         `json:"env_group" orm:"column(env_group);size(20);null(false);description(环境组)"`
	ComponentID     string         `json:"component_id" orm:"column(component_id);size(36);null;description(组件ID)"`
	Description     string         `json:"description" orm:"column(description);type(text);null;description(环境描述)"`
	ImagesAddr      string         `json:"images_addr" orm:"column(images_addr);size(255);null;description(镜像地址)"`
	ImagesUser      string         `json:"images_user" orm:"column(images_user);size(255);null;description(镜像用户)"`
	ImagesPwd       string         `json:"images_pwd" orm:"column(images_pwd);size(255);null;description(镜像密码)"`
	KubernetesAddr  string         `json:"kubernetes_addr" orm:"column(kubernetes_addr);size(255);null;description(kubernetes地址)"`
	Kubeconfig      string         `json:"kubeconfig" orm:"column(kubeconfig);type(text);null;description(kubeconfig)"`
	KubeNamespace   string         `json:"kube_namespace" orm:"column(kube_namespace);size(255);null;description(kubernetes命名空间)"`
	ComponentValues string         `json:"component_values" orm:"column(component_values);type(text);null;description(组件构建参数)"`
	CreatedAt       time.Time      `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt       time.Time      `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
	Children        []*Environment `json:"children" orm:"-"`
}

// 服务树为空时，添加一条默认数据
func DefaultEnvTableString() *Environment {
	return &Environment{
		Id:             uuid.New().String(),
		Name:           "Default",
		IsProd:         false,
		IsEnv:          false,
		EnvGroup:       "",
		Description:    "Default",
		ImagesAddr:     "",
		ImagesUser:     "",
		ImagesPwd:      "",
		KubernetesAddr: "",
		Kubeconfig:     "",
		KubeNamespace:  "",
	}
}
