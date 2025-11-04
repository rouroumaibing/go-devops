package models

import (
	"fmt"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/utils/hash"
)

func init() {
	orm.RegisterModel(new(auths.Users))
}

func GetUserInfo() ([]*auths.Users, error) {
	lock.RLock()
	defer lock.RUnlock()
	var users []*auths.Users
	_, err := orm.NewOrm().QueryTable("users").All(&users)
	if err != nil {
		return nil, err
	}
	// 不返回密码
	for _, user := range users {
		user.Password = ""
	}

	return users, nil
}

func GetUserInfoByID(useruuid string) (*auths.Users, error) {
	lock.RLock()
	defer lock.RUnlock()

	var user auths.Users
	err := orm.NewOrm().QueryTable("users").Filter("id", useruuid).One(&user)
	if err != nil {
		return nil, err
	}
	// 不返回密码
	user.Password = ""

	return &user, nil
}

func UpdateUserInfoByID(useruuid string, user *auths.Users) error {
	sql := "UPDATE users SET updated_at = CURRENT_TIMESTAMP"
	params := []interface{}{}
	fields := []string{}

	if user.Accountname != "" {
		fields = append(fields, "accountname = ?")
		params = append(params, user.Accountname)
	}
	if user.Accountgroup != "" {
		fields = append(fields, "accountgroup = ?")
		params = append(params, user.Accountgroup)
	}
	if user.Nickname != "" {
		fields = append(fields, "nickname = ?")
		params = append(params, user.Nickname)
	}
	if user.HeadImg != "" {
		fields = append(fields, "head_img = ?")
		params = append(params, user.HeadImg)
	}
	if user.Age > 0 {
		fields = append(fields, "age = ?")
		params = append(params, user.Age)
	}
	if user.Address != "" {
		fields = append(fields, "address = ?")
		params = append(params, user.Address)
	}
	if user.Password != "" {
		// 加密密码
		user.Password = hash.MD5(user.Password)
		fields = append(fields, "password = ?")
		params = append(params, user.Password)
	}
	if user.Email != "" {
		fields = append(fields, "email = ?")
		params = append(params, user.Email)
	}
	if user.Phone != "" {
		fields = append(fields, "phone = ?")
		params = append(params, user.Phone)
	}
	if user.Wechat != "" {
		fields = append(fields, "wechat = ?")
		params = append(params, user.Wechat)
	}
	if user.Qq != "" {
		fields = append(fields, "qq = ?")
		params = append(params, user.Qq)
	}
	if user.Identitycard != "" {
		fields = append(fields, "identitycard = ?")
		params = append(params, user.Identitycard)
	}

	// 没有需要更新的字段
	if len(fields) == 0 {
		return nil
	}

	sql += ", " + strings.Join(fields, ", ") + " WHERE id = ?"
	params = append(params, useruuid)

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	res, err := dao.Raw(sql, params...).Exec()
	if err != nil {
		return fmt.Errorf("更新用户信息失败: %v", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil || rowsAffected == 0 {
		return fmt.Errorf("未找到ID为%s的用户或没有字段被更新", useruuid)
	}

	return nil
}

func DeleteUserInfoByID(useruuid string) error {
	dao := orm.NewOrm()
	_, err := dao.Delete(&auths.Users{Id: useruuid})
	if err != nil {
		return err
	}
	return nil
}
