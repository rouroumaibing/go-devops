package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
)

func init() {
	orm.RegisterModel(new(service.ChangeLog))

}

func GetChangeById(id string) (*service.ChangeLog, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `SELECT id, component_id, change_number, description, content, status, creator, last_update_time FROM change_log WHERE id = ?`
	var change service.ChangeLog
	err := dao.Raw(sql, id).QueryRow(&change)
	if err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("找不到变更记录,ID: %s", id)
		}
		return nil, err
	}
	return &change, nil
}

func CreateChange(change *service.ChangeLog) error {

	if change.ComponentId == "" || change.ChangeNumber == "" || change.Creator == "" {
		return fmt.Errorf("组件ID、变更编号和创建人不能为空")
	}

	lock.Lock()
	defer lock.Unlock()

	changeID := uuid.New().String()
	dao := orm.NewOrm()
	sql := `INSERT INTO change_log (id, component_id, change_number, description, content, status, creator, last_update_time) VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`

	_, err := dao.Raw(sql, changeID, change.ComponentId, change.ChangeNumber, change.Description, change.Content, change.Status, change.Creator).Exec()
	if err != nil {
		return fmt.Errorf("创建变更失败: %v", err)
	}
	return nil
}

func UpdateChangeById(change *service.ChangeLog) error {
	sql := "UPDATE change_log SET last_update_time = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if change.ComponentId != "" {
		fields = append(fields, "component_id = ?")
		params = append(params, change.ComponentId)
	}
	if change.ChangeNumber != "" {
		fields = append(fields, "change_number = ?")
		params = append(params, change.ChangeNumber)
	}
	if change.Description != "" {
		fields = append(fields, "description = ?")
		params = append(params, change.Description)
	}
	if change.Content != "" {
		fields = append(fields, "content = ?")
		params = append(params, change.Content)
	}
	if change.Status != "" {
		fields = append(fields, "status = ?")
		params = append(params, change.Status)
	}
	if change.Creator != "" {
		fields = append(fields, "creator = ?")
		params = append(params, change.Creator)
	}

	// 没有需要更新的字段
	if len(fields) == 0 {
		return nil
	}

	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, change.Id)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新变更失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的变更或没有字段被更新", change.Id)
	}

	return nil
}

func DeleteChangeById(id string) error {
	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	sql := `DELETE FROM change_log WHERE id = ?`
	_, err := dao.Raw(sql, id).Exec()
	if err != nil {
		return fmt.Errorf("删除变更失败: %v", err)
	}
	return nil
}
