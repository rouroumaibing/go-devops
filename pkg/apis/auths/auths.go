package auths

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	AccessSecret  = []byte("ACCESS#2025@09!09")
	RefreshSecret = []byte("REFRESH#2025@09~09")
)

var RequestURIWhiteList = []string{
	"/api/version",
	"/api/auth/register",
	"/api/auth/refresh",
	"/api/auth/login",
	"/api/auth/sms/login",
	"/api/auth/wechat/check",
	"/api/auth/wechat/callback",
	"/api/auth/qq/callback",
}

const (
	JWTIssuer          = "IAM"
	TokenExpire        = 24 * time.Hour
	AccessTokenExpire  = 15 * time.Minute
	RefreshTokenExpire = 7 * 24 * time.Hour
	SMSExpire          = 5 * time.Minute
)

type WeChatCallbackRequest struct {
	Code string `json:"code"`
}

type SMS struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type JWTClaims struct {
	UserID string `json:"id"`
	jwt.RegisteredClaims
}

type TokenResponse struct {
	UserUUID     string `json:"useruuid"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

// mysql判断空值也是唯一
type Users struct {
	Id           string    `json:"id" orm:"pk;size(36);unique;description(主键ID)"`
	Accountname  string    `json:"accountname" orm:"size(64);unique;description(账户名)"`
	Accountgroup string    `json:"accountgroup" orm:"size(64);description(账户组)"`
	Nickname     string    `json:"nickname" orm:"size(64);description(昵称)"`
	HeadImg      string    `json:"head_img" orm:"size(1024);description(头像)"`
	Age          int       `json:"age" orm:"size(64);description(年龄)"`
	Address      string    `json:"address" orm:"size(1024);description(地址)"`
	Password     string    `json:"password" orm:"size(1024);description(密码)"`
	Email        string    `json:"email" orm:"size(64);unique;description(邮箱)"`
	Phone        string    `json:"phone" orm:"size(11);unique;description(电话)"`
	Wechat       string    `json:"wechat" orm:"size(64);unique;description(微信)"`
	Qq           string    `json:"qq" orm:"size(64);unique;description(QQ)"`
	Identitycard string    `json:"identitycard" orm:"size(18);unique;description(身份证)"`
	CreatedAt    time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt    time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}
