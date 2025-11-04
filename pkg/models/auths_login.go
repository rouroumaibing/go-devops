package models

import (
	"encoding/json"
	"fmt"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	jwtauth "github.com/rouroumaibing/go-devops/pkg/utils/auth/jwt"
	oauth2_0 "github.com/rouroumaibing/go-devops/pkg/utils/auth/oauth"
	redisclient "github.com/rouroumaibing/go-devops/pkg/utils/db"
	"github.com/rouroumaibing/go-devops/pkg/utils/hash"
)

func NewToken(useruuid string) (string, string, error) {
	accessToken, refreshToken, err := jwtauth.GenerateTokens(useruuid)
	if err != nil {
		return "", "", err
	}

	_, err = redisclient.Dao("set", useruuid+"_refresh_token", refreshToken, auths.RefreshTokenExpire.String())
	if err != nil {
		return "", "", err
	}

	_, err = redisclient.Dao("set", useruuid+"_access_token", accessToken, auths.AccessTokenExpire.String())
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func CheckAccessToken(token string) (string, error) {
	claims, err := jwtauth.ParseAccessToken(token)
	if err != nil {
		return "", err
	}

	return claims.UserID, nil
}

func CheckRefreshToken(token string) (string, error) {
	claims, err := jwtauth.ParseRefreshToken(token)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

func RefreshUserToken(useruuid, refreshToken string) (string, error) {
	accessToken, _, err := jwtauth.RefreshUserToken(refreshToken)
	if err != nil {
		return "", err
	}
	_, err = redisclient.Dao("set", useruuid+"_access_token", accessToken, auths.AccessTokenExpire.String())
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

func DeleteAccessToken(useruuid string) error {
	// 从redis中删除Token
	_, err := redisclient.Dao("del", useruuid+"_access_token")
	if err != nil {
		return err
	}
	return nil
}

func LoginWithPwd(user *auths.Users) (*auths.Users, error) {
	// 保存输入密码的MD5值
	inputPasswordHash := hash.MD5(user.Password)

	lock.RLock()
	defer lock.RUnlock()

	// 修改校验方法为查询出sql里的密码
	// 与输入的密码的md5值进行比较
	dao := orm.NewOrm()
	sql := `SELECT id, accountname, nickname, head_img, age, address, email, phone, wechat, qq, identitycard, password, created_at, updated_at FROM users WHERE accountname = ?`

	// 创建一个新的变量来存储查询结果，而不是复用输入的user
	dbUser := &auths.Users{}
	err := dao.Raw(sql, user.Accountname).QueryRow(dbUser)
	if err != nil {
		return nil, fmt.Errorf("登录失败: %v", err)
	}

	// 校验密码：比较输入密码的MD5与数据库中存储的密码
	if inputPasswordHash != dbUser.Password {
		return nil, fmt.Errorf("登录失败: 密码错误")
	}

	// 不返回密码字段给前端
	dbUser.Password = ""

	return dbUser, nil
}

func WeChatCallback(code string) (string, string, string, error) {
	// 1. 获取token URL
	tokenURL := oauth2_0.WeChatCallback(code)

	// 2. 发送请求获取access_token
	tokenResp, err := oauth2_0.GetWechatAccessToken(tokenURL)
	if err != nil {
		return "", "", "", err
	}

	// 3. 使用access_token获取用户信息
	userInfo, err := oauth2_0.GetWechatUserInfo(tokenResp.AccessToken, tokenResp.OpenID)
	if err != nil {
		return "", "", "", err
	}

	// 4. 构建用户会话信息
	sessionInfo := &auths.WechatUserSessionInfo{
		OpenID:     userInfo.OpenID,
		Nickname:   userInfo.Nickname,
		HeadImg:    userInfo.HeadImg,
		UnionID:    userInfo.UnionID,
		IsLoggedIn: true,
	}

	// 5. 将用户会话信息序列化为JSON并存储到Redis
	sessionJSON, err := json.Marshal(sessionInfo)
	if err != nil {
		return "", "", "", fmt.Errorf("序列化用户会话信息失败: %v", err)
	}

	// 使用OpenID作为Redis的键，存储用户会话信息
	_, err = redisclient.Dao("set", userInfo.OpenID, string(sessionJSON), auths.TokenExpire.String())
	if err != nil {
		return "", "", "", fmt.Errorf("存储用户会话信息到Redis失败: %v", err)
	}

	// 6. 可以添加用户信息存储到数据库的逻辑
	// 检查用户是否存在，如果不存在则创建新用户
	user, err := FindUserByWechat(userInfo.OpenID)
	if err != nil {
		return "", "", "", fmt.Errorf("查询用户失败: %v", err)
	}
	if user == nil {
		// 创建新用户
		user = &auths.Users{
			Nickname: userInfo.Nickname,
			Wechat:   userInfo.OpenID,
			HeadImg:  userInfo.HeadImg,
		}
		user.Id = uuid.New().String()

		lock.Lock()
		defer lock.Unlock()

		_, err = orm.NewOrm().Insert(user)
		if err != nil {
			return "", "", "", fmt.Errorf("创建用户失败: %v", err)
		}
	}

	// 7.设置token
	accessToken, refreshToken, err := NewToken(user.Id)
	if err != nil {
		return "", "", "", fmt.Errorf("创建token失败: %v", err)
	}

	return user.Id, accessToken, refreshToken, nil
}
