package models

import (
	"fmt"
	"strings"

	"k8s.io/klog/v2"

	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/utils/auth"
	redisclient "github.com/rouroumaibing/go-devops/pkg/utils/db"
)

func SendSMS(phone string) bool {
	code := auth.GenerateCode()
	// 调用三方SDK
	klog.Info("调用第三方SDK API发送短信")

	fmt.Println("code:", code)

	success := SaveCodeToCache(phone, code)
	return success
}

func SaveCodeToCache(phone string, code string) bool {
	_, err := redisclient.Dao("set", phone, code, auths.SMSExpire.String())
	if err != nil {
		klog.Error("保存验证码失败：", phone, ":", err)
		return false
	}

	return true
}

func ValidateCode(phone string, code string) bool {
	val, err := redisclient.Dao("get", phone)
	if err != nil {
		klog.Error("验证码查询失败：", phone, ":", err)
		return false
	}

	result := strings.Compare(val, code)
	if result != 0 {
		klog.Error("验证码不同：", phone, ":", err)
		return false
	}

	_, err = redisclient.Dao("del", phone)
	if err != nil {
		klog.Error("删除验证码失败：", phone, ":", err)
		return false
	}

	return true
}

func ValidateUserExists(phone string) (string, error) {
	lock.RLock()
	defer lock.RUnlock()

	dao := orm.NewOrm()
	sql := `SELECT id FROM users WHERE phone = ?`

	var id string
	err := dao.Raw(sql, phone).QueryRow(&id)
	if err != nil {
		return "", fmt.Errorf("用户不存在: %v", err)
	}

	return id, nil
}
