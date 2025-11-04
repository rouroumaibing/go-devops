package models

import (
	"fmt"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/servicetree"
)

func init() {
	orm.RegisterModel(new(servicetree.ServiceTree))
}

var lock sync.RWMutex

func GetServiceTree() ([]*servicetree.ServiceTree, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `select * from service_tree`
	var serviceTree []*servicetree.ServiceTree
	_, err := dao.Raw(sql).QueryRows(&serviceTree)
	if err != nil {
		fmt.Printf("查询服务树失败: %v", err)
		return nil, err
	}

	if len(serviceTree) == 0 {
		addsql := `insert ignore into service_tree(id, name, full_path, node_type, service_id, parent_id, level, description, created_at, updated_at) 
		values(?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
		defaultServiceTree := servicetree.DefaultServiceTreeTableString()
		_, err := dao.Raw(addsql,
			defaultServiceTree.Id,
			defaultServiceTree.Name,
			defaultServiceTree.FullPath,
			defaultServiceTree.NodeType,
			defaultServiceTree.ServiceId,
			defaultServiceTree.ParentId,
			defaultServiceTree.Level,
			defaultServiceTree.Description,
		).Exec()
		if err != nil {
			return nil, fmt.Errorf("插入默认数据失败: %v", err)
		}
		_, err = dao.Raw(sql).QueryRows(&serviceTree)
		if err != nil {
			return nil, err
		}
	}
	resultServiceTree := buildTree(serviceTree)

	return resultServiceTree, nil
}

func buildTree(serviceTree []*servicetree.ServiceTree) []*servicetree.ServiceTree {
	// 创建节点ID到节点的映射表，便于快速查找父节点
	nodeMap := make(map[string]*servicetree.ServiceTree)
	for _, node := range serviceTree {
		// 初始化子节点切片
		node.Children = []*servicetree.ServiceTree{}
		nodeMap[node.Id] = node
	}

	// 构建树形结构
	var roots []*servicetree.ServiceTree
	for _, node := range serviceTree {
		if node.ParentId == "root" {
			roots = append(roots, node)
		} else {
			// 查找父节点并添加当前节点为子节点
			if parent, exists := nodeMap[node.ParentId]; exists {
				parent.Children = append(parent.Children, node)
			}
		}
	}

	return roots
}

func GetServiceTreeById(id string) (*servicetree.ServiceTree, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	// 显式指定所有字段，避免*带来的字段顺序依赖问题
	sql := `SELECT id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at 
		FROM service_tree WHERE id = ?`
	var serviceTree servicetree.ServiceTree
	err := dao.Raw(sql, id).QueryRow(&serviceTree)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("服务树节点不存在,ID: %s", id)
		}
		return nil, fmt.Errorf("查询服务树节点失败: %v", err)
	}
	return &serviceTree, nil
}
func CreateServiceTree(serviceTree *servicetree.ServiceTree) error {
	lock.Lock()
	defer lock.Unlock()

	servicetreeID := uuid.New().String()
	dao := orm.NewOrm()
	sql := `insert into service_tree(id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at) 
		values(?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	_, err := dao.Raw(sql,
		servicetreeID,
		serviceTree.Name,
		serviceTree.FullPath,
		serviceTree.NodeType,
		serviceTree.ParentId,
		serviceTree.Level,
		serviceTree.ServiceId,
		serviceTree.Description,
	).Exec()
	if err != nil {
		return err
	}
	return nil
}
func UpdateServiceTreeById(serviceTree *servicetree.ServiceTree) error {
	// 构建动态更新SQL
	sql := "UPDATE service_tree SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if serviceTree.Name != "" {
		fields = append(fields, "name = ?")
		params = append(params, serviceTree.Name)
	}

	if serviceTree.FullPath != "" {
		fields = append(fields, "full_path = ?")
		params = append(params, serviceTree.FullPath)
	}

	if serviceTree.NodeType != "" {
		fields = append(fields, "node_type = ?")
		params = append(params, serviceTree.NodeType)
	}

	// 第一层级节点ID为root，不更新该字段
	if serviceTree.ParentId != "" && serviceTree.ParentId != "root" {
		fields = append(fields, "parent_id = ?")
		params = append(params, serviceTree.ParentId)
	}

	if serviceTree.Level > 0 {
		fields = append(fields, "level = ?")
		params = append(params, serviceTree.Level)
	}

	if serviceTree.ServiceId != "" {
		fields = append(fields, "service_id = ?")
		params = append(params, serviceTree.ServiceId)
	}

	if serviceTree.Description != "" {
		fields = append(fields, "description = ?")
		params = append(params, serviceTree.Description)
	}

	// 如果没有需要更新的字段
	if len(fields) == 0 {
		return nil
	}

	// 拼接SQL语句
	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, serviceTree.Id)

	lock.Lock()
	defer lock.Unlock()

	// 执行更新
	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新服务失败: %v", err)
	}

	// 检查是否有记录被更新
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新行数失败: %v", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的服务或没有字段被更新", serviceTree.Id)
	}

	return nil
}

func DeleteServiceTreeById(id string) error {
	isHaveNode, err := CheckServiceTreeId(id)
	if err != nil {
		return err
	}
	if isHaveNode {
		return fmt.Errorf("服务树节点下存在子节点或微服务，不允许删除")
	}

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `delete from service_tree where id = ?`
	_, err = dao.Raw(sql, id).Exec()
	if err != nil {
		return err
	}
	return nil
}

// 新增一个函数，检查服务树ID下是否还存在节点或微服务
func CheckServiceTreeId(id string) (bool, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `select count(*) from service_tree where parent_id = ?`
	var count int
	err := dao.Raw(sql, id).QueryRow(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
