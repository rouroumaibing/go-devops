package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

func init() {
	orm.RegisterModel(new(service.Environment))
}

func BuildEnvironmentTree(environments []*service.Environment) ([]*service.Environment, error) {
	if len(environments) == 0 {
		return []*service.Environment{}, nil
	}

	// 创建映射，按名称索引
	envByName := make(map[string]*service.Environment)

	// 初始化所有节点的Children为空切片
	for _, env := range environments {
		envByName[env.Name] = env
		env.Children = []*service.Environment{}
	}

	// 记录哪些节点是子节点
	isChild := make(map[string]bool)

	// 构建树形结构 - 将结束节点（IsEnv=true）添加到其父节点（IsEnv=false）的children中
	for _, env := range environments {
		// 跳过没有父节点的节点
		if env.EnvGroup == "" || env.EnvGroup == env.Name {
			continue
		}

		// 查找父节点（通过名称匹配）
		if parent, exists := envByName[env.EnvGroup]; exists {
			// 确保父节点不是结束节点（IsEnv=false）
			// 结束节点（IsEnv=true）可以作为子节点
			if !parent.IsEnv {
				parent.Children = append(parent.Children, env)
				isChild[env.Name] = true
			}
		}
	}

	// 返回根节点（没有被其他节点作为子节点的节点）
	var result []*service.Environment
	for _, env := range environments {
		if !isChild[env.Name] {
			result = append(result, env)
		}
	}

	return result, nil
}

func SplitEnvironmentTree(environments []*service.Environment) ([]*service.Environment, error) {
	if len(environments) == 0 {
		return []*service.Environment{}, nil
	}
	var envList []*service.Environment
	for _, env := range environments {
		envList = append(envList, env)
		if len(env.Children) > 0 {
			children, err := SplitEnvironmentTree(env.Children)
			if err != nil {
				return nil, err
			}
			envList = append(envList, children...)
		}
	}
	return envList, nil
}

func GetEnvironmentById(id string) (*service.Environment, error) {
	lock.RLock()
	defer lock.RUnlock()

	var environment service.Environment
	err := orm.NewOrm().QueryTable("environment").Filter("id", id).One(&environment)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("环境不存在,ID: %s", id)
		}
		return nil, fmt.Errorf("查询环境失败: %v", err)
	}

	return &environment, nil
}

// 单个单个节点添加
func CreateEnvironment(e *service.Environment) error {
	if e.Name == "" {
		return fmt.Errorf("环境名称不能为空")
	}

	lock.Lock()
	defer lock.Unlock()

	e.Id = uuid.New().String()
	_, err := orm.NewOrm().Insert(e)
	if err != nil {
		return fmt.Errorf("创建环境失败: %v", err)
	}
	return nil
}

func UpdateEnvironmentById(e *service.Environment) error {
	sql := "UPDATE environment SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if e.Name != "" {
		fields = append(fields, "name = ?")
		params = append(params, e.Name)
	}

	fields = append(fields, "is_prod = ?")
	params = append(params, e.IsProd)

	fields = append(fields, "is_env = ?")
	params = append(params, e.IsEnv)

	if e.EnvGroup != "" {
		fields = append(fields, "env_group = ?")
		params = append(params, e.EnvGroup)
	}
	if e.ComponentID != "" {
		fields = append(fields, "component_id = ?")
		params = append(params, e.ComponentID)
	}
	if e.Description != "" {
		fields = append(fields, "description = ?")
		params = append(params, e.Description)
	}
	if e.ImagesAddr != "" {
		fields = append(fields, "images_addr = ?")
		params = append(params, e.ImagesAddr)
	}
	if e.ImagesUser != "" {
		fields = append(fields, "images_user = ?")
		params = append(params, e.ImagesUser)
	}
	if e.ImagesPwd != "" {
		fields = append(fields, "images_pwd = ?")
		params = append(params, e.ImagesPwd)
	}
	if e.KubernetesAddr != "" {
		fields = append(fields, "kubernetes_addr = ?")
		params = append(params, e.KubernetesAddr)
	}
	if e.Kubeconfig != "" {
		fields = append(fields, "kubeconfig = ?")
		params = append(params, e.Kubeconfig)
	}
	if e.KubeNamespace != "" {
		fields = append(fields, "kube_namespace = ?")
		params = append(params, e.KubeNamespace)
	}
	if e.ComponentValues != "" {
		fields = append(fields, "component_values = ?")
		params = append(params, e.ComponentValues)
	}

	if len(fields) == 0 {
		return nil
	}

	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, e.Id)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新环境失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的环境或没有字段被更新", e.Id)
	}

	return nil
}
func DeleteEnvironmentById(id string) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `DELETE FROM environment WHERE id = ?`
	_, err := dao.Raw(sql, id).Exec()
	if err != nil {
		return fmt.Errorf("删除环境失败: %v", err)
	}
	return nil
}
