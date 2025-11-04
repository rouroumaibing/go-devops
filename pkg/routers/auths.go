package routers

import (
	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/controllers/auths"
)

func init() {
	LoginRouter()
	RegisterRouter()
	SMSRouter()
	UserInfoRouter()
}

func LoginRouter() {
	beego.Router("/api/auth/refresh", &auths.LoginController{}, "post:RefreshToken")
	beego.Router("/api/auth/login", &auths.LoginController{}, "post:LoginPWD")
	beego.Router("/api/auth/logout", &auths.LoginController{}, "post:Logout")
	beego.Router("/api/auth/wechat/check", &auths.LoginController{}, "post:WeChatCheck")
	beego.Router("/api/auth/wechat/callback", &auths.LoginController{}, "post:WeChatCallback")
	beego.Router("/api/auth/qq/callback", &auths.LoginController{}, "post:QQCallback")
}

func RegisterRouter() {
	beego.Router("/api/auth/register", &auths.RegisterController{}, "post:Register")
}

func UserInfoRouter() {
	beego.Router("/api/auth/users", &auths.UserInfoController{}, "get:GetUserInfo")
	beego.Router("/api/auth/users/:id", &auths.UserInfoController{}, "get:GetUserInfoByID")
	beego.Router("/api/auth/users/:id", &auths.UserInfoController{}, "put:UpdateUserInfoByID")
	beego.Router("/api/auth/users/:id", &auths.UserInfoController{}, "delete:DeleteUserInfoByID")
}

func SMSRouter() {
	beego.Router("/api/auth/sms/login", &auths.SMSController{}, "post:SMSLogin")
	beego.Router("/api/auth/sms", &auths.SMSController{}, "get:GetSMS")
}
