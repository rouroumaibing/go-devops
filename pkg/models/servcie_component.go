package models

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

func init() {
	orm.RegisterModel(new(service.Component))
}

// Component相关方法
func GetComponent() ([]*service.Component, error) {
	var components []*service.Component
	_, err := orm.NewOrm().QueryTable(new(service.Component)).All(&components)
	if err != nil {
		return nil, err
	}
	return components, nil
}

func GetComponentById(id string) (*service.Component, error) {
	var component service.Component
	err := orm.NewOrm().QueryTable(new(service.Component)).Filter("id", id).One(&component)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("组件不存在,ID: %s", id)
		}
		return nil, fmt.Errorf("查询组件失败: %v", err)
	}
	return &component, nil
}

func CreateComponent(component *service.Component) ([]byte, error) {
	if component.Name == "" || component.RepoUrl == "" || component.RepoBranch == "" || component.Owner == "" {
		return []byte(""), fmt.Errorf("组件名称、代码库地址、分支和责任人不能为空")
	}
	lock.Lock()
	defer lock.Unlock()

	componentID := uuid.New().String()
	dao := orm.NewOrm()
	sql := `INSERT INTO component (id, name, service_tree, owner, description, repo_url, repo_branch, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err := dao.Raw(sql, componentID, component.Name, component.ServiceTree, component.Owner, component.Description, component.RepoUrl, component.RepoBranch).Exec()
	if err != nil {
		return []byte(""), fmt.Errorf("创建组件失败: %v", err)

	}

	componentIdJson := map[string]string{
		"id": componentID,
	}
	componentIdJsonBytes, err := json.Marshal(componentIdJson)
	if err != nil {
		return []byte(""), fmt.Errorf("组件ID转换为JSON失败: %v", err)
	}

	return componentIdJsonBytes, nil
}

func UpdateComponentById(c *service.Component) error {
	sql := "UPDATE component SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if c.Name != "" {
		fields = append(fields, "name = ?")
		params = append(params, c.Name)
	}
	if c.ServiceTree != "" {
		fields = append(fields, "service_tree = ?")
		params = append(params, c.ServiceTree)
	}
	if c.Owner != "" {
		fields = append(fields, "owner = ?")
		params = append(params, c.Owner)
	}
	if c.Description != "" {
		fields = append(fields, "description = ?")
		params = append(params, c.Description)
	}
	if c.RepoUrl != "" {
		fields = append(fields, "repo_url = ?")
		params = append(params, c.RepoUrl)
	}
	if c.RepoBranch != "" {
		fields = append(fields, "repo_branch = ?")
		params = append(params, c.RepoBranch)
	}
	// 没有需要更新的字段
	if len(fields) == 0 {
		return nil
	}

	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, c.Id)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新组件失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的组件或没有字段被更新", c.Id)
	}

	return nil
}
func DeleteComponentById(id string) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `DELETE FROM component WHERE id = ?`
	_, err := dao.Raw(sql, id).Exec()
	if err != nil {
		return fmt.Errorf("删除组件失败: %v", err)
	}
	return nil
}

func GetComponentPipelines(id string) ([]*service.Pipeline, error) {
	var pipelines []*service.Pipeline

	_, err := orm.NewOrm().QueryTable(new(service.Pipeline)).Filter("component_id", id).All(&pipelines)
	if err != nil {
		return nil, fmt.Errorf("查询组件流水线失败: %v", err)
	}

	// 2. 对于每个流水线，获取其阶段信息
	for _, pipeline := range pipelines {
		var stages []*service.PipelineStage
		_, err = orm.NewOrm().QueryTable(new(service.PipelineStage)).Filter("pipeline_id", pipeline.Id).All(&stages)
		if err != nil {
			return nil, fmt.Errorf("查询流水线阶段失败: %v", err)
		}
		pipeline.PipelineStages = stages
	}

	return pipelines, nil
}

func GetComponentEnvs(id string) ([]*service.Environment, error) {
	lock.RLock()
	defer lock.RUnlock()

	var envs []*service.Environment
	_, err := orm.NewOrm().QueryTable("environment").Filter("component_id", id).All(&envs)

	if err != nil {
		return nil, fmt.Errorf("查询组件环境失败: %v", err)
	}

	// 如果没有找到该组件的环境，插入默认环境
	if len(envs) == 0 {
		defaultEnv := service.DefaultEnvTableString()
		// 插入默认的一条数据
		_, err = orm.NewOrm().Insert(&defaultEnv)
		if err != nil {
			return nil, fmt.Errorf("插入默认环境失败: %v", err)
		}
		// 重新查询插入的环境
		_, err = orm.NewOrm().QueryTable("environment").Filter("component_id", id).All(&envs)
		if err != nil {
			return nil, fmt.Errorf("查询新插入的环境失败: %v", err)
		}
	}

	envs, err = BuildEnvironmentTree(envs)
	if err != nil {
		return nil, err
	}
	return envs, nil
}

func GetComponentProducts(id string) ([]*service.Product, error) {
	var products []*service.Product
	_, err := orm.NewOrm().QueryTable(new(service.Product)).Filter("component_id", id).All(&products)
	if err != nil {
		return nil, fmt.Errorf("查询组件产品失败: %v", err)
	}

	return products, nil
}
func GetComponentChanges(id string) ([]*service.ChangeLog, error) {
	var changes []*service.ChangeLog
	_, err := orm.NewOrm().QueryTable(new(service.ChangeLog)).Filter("component_id", id).All(&changes)
	if err != nil {
		return nil, fmt.Errorf("查询组件变更失败: %v", err)
	}
	return changes, nil
}

func CreateComponentParameter(parameter *service.ComponentParameter) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	parameter.Id = uuid.New().String()
	_, err := dao.Insert(parameter)
	if err != nil {
		return fmt.Errorf("创建组件参数失败: %v", err)
	}
	return nil
}
func GetComponentParameterById(componentID string, environmentID string) (*service.ComponentParameter, error) {
	lock.RLock()
	defer lock.RUnlock()

	var parameter service.ComponentParameter
	err := orm.NewOrm().QueryTable(new(service.ComponentParameter)).Filter("component_id", componentID).Filter("environment_id", environmentID).One(&parameter)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("组件参数不存在: %w", err)
		}
		return nil, fmt.Errorf("查询组件参数失败: %w", err)
	}
	return &parameter, nil
}
func UpdateComponentParameterById(componentID string, environmentID string, parameter *service.ComponentParameter) error {
	sql := "UPDATE component_parameter SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}
	if parameter.ParamYaml != "" {
		fields = append(fields, "param_yaml = ?")
		params = append(params, parameter.ParamYaml)
	}

	if len(fields) == 0 {
		return nil
	}
	sql += ", " + strings.Join(fields, ", ") + " WHERE component_id = ? AND environment_id = ?"
	params = append(params, componentID, environmentID)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新组件参数失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的组件参数或没有字段被更新", parameter.Id)
	}

	return nil
}

func DeleteComponentParameterById(componentID string, environmentID string) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `DELETE FROM component_parameter WHERE component_id = ? AND environment_id = ?`
	_, err := dao.Raw(sql, componentID, environmentID).Exec()
	if err != nil {
		return fmt.Errorf("删除组件参数失败: %v", err)
	}
	return nil
}
