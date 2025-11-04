package servicetree

import (
	"time"

	"github.com/google/uuid"
)

const ()

// ServiceTree API层服务树结构体，合并数据库模型与API响应格式
type ServiceTree struct {
	Id          string         `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name        string         `json:"name" orm:"column(name);size(100);null(false);description(节点名称)"`
	FullPath    string         `json:"full_path" orm:"column(full_path);size(500);null(false);unique;description(完整路径)"`
	NodeType    string         `json:"node_type" orm:"column(node_type);size(20);null(false);default(category);description(节点类型:category/subcategory/service)"`
	ParentId    string         `json:"parent_id,omitempty" orm:"column(parent_id);type(string unsigned);null;description(父节点ID)"`
	Level       int            `json:"level" orm:"column(level);type(tinyint unsigned);null(false);default(1);description(层级深度)"`
	ServiceId   string         `json:"service_id,omitempty" orm:"column(service_id);type(string unsigned);null;description(服务ID)"`
	Description string         `json:"description,omitempty" orm:"column(description);type(text);null;description(节点描述)"`
	CreatedAt   *time.Time     `json:"created_at" orm:"column(created_at);auto_now_add;type(timestamp);description(创建时间)"`
	UpdatedAt   *time.Time     `json:"updated_at" orm:"column(updated_at);auto_now;type(timestamp);description(更新时间)"`
	Children    []*ServiceTree `json:"children,omitempty" orm:"-"` // 添加子节点字段，orm:"-"表示不映射数据库
}

// 服务树为空时，添加一条默认数据
func DefaultServiceTreeTableString() *ServiceTree {
	return &ServiceTree{
		Id:          uuid.New().String(),
		Name:        "Default",
		FullPath:    "Default",
		NodeType:    "category",
		ParentId:    "root", // root代表第一层层级
		Level:       1,
		ServiceId:   "",
		Description: "Default",
	}
}
