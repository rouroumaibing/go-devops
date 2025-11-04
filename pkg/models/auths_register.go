package models

import (
	"errors"
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/utils/hash"
)

func RegisterUserWithPWD(user *auths.Users) (err error) {
	tmpUser, err := FindUserByName(user.Accountname)
	// 查询到用户，返回错误
	if tmpUser != nil {
		return errors.New("accountname already exists")
	}
	// 查询失败，且不是没有数据错误，返回错误
	if err != nil && err != orm.ErrNoRows {
		return err
	}

	lock.Lock()
	defer lock.Unlock()

	user.Password = hash.MD5(user.Password)
	user.Id = uuid.New().String()

	_, err = orm.NewOrm().Insert(user)
	if err != nil {
		return fmt.Errorf("注册用户失败: %v", err)
	}
	return nil
}

func FindUserByName(name string) (user *auths.Users, err error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `SELECT id, accountname, nickname, head_img, age, address, email, phone, wechat, qq, identitycard, created_at, updated_at FROM users WHERE accountname = ?`

	user = &auths.Users{}
	err = dao.Raw(sql, name).QueryRow(user)
	if err != nil {
		// 如果没有找到用户，返回nil和nil
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}
	return user, nil
}

func FindUserByPhone(phone string) (user *auths.Users, err error) {
	lock.RLock()
	defer lock.RUnlock()

	user = &auths.Users{}
	err = orm.NewOrm().QueryTable("users").Filter("Phone", phone).One(&user)
	if err != nil {
		// 如果没有找到用户，返回nil和nil
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return user, nil
}

func FindUserByWechat(wechat string) (user *auths.Users, err error) {
	lock.RLock()
	defer lock.RUnlock()

	user = &auths.Users{}
	err = orm.NewOrm().QueryTable("users").Filter("Wechat", wechat).One(&user)
	if err != nil {
		// 如果没有找到用户，返回nil和nil
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("查询用户失败: %v", err)
	}

	return user, nil
}
