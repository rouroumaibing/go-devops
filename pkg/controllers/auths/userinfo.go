package auths

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

// 用户的增删改查
type UserInfoController struct {
	beego.Controller
}

func (this *UserInfoController) GetUserInfo() {
	// 从数据库中获取用户信息
	userinfo, err := models.GetUserInfo()
	if err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.User.QueryFailed.WithMessage("get userinfo from db faild!"))
		return
	}
	userJson, err := json.Marshal(userinfo)
	if err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal userinfo faild!"))
		return
	}
	// 响应
	response.HttpResponse(http.StatusOK, userJson, this.Ctx)
}

// 获取用户信息
func (this *UserInfoController) GetUserInfoByID() {
	userID := this.Ctx.Input.Param(":id")
	if userID == "" {
		response.SetErrorResponse(this.Ctx, errcode.E.User.ParamError.WithMessage("get use id failed!"))
		return
	}
	// 从数据库中获取用户信息
	userinfo, err := models.GetUserInfoByID(userID)
	if err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.User.QueryFailed.WithMessage("get userinfo from db faild!"))
		return
	}
	userJson, err := json.Marshal(userinfo)
	if err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal userinfo faild!"))
		return
	}
	// 响应
	response.HttpResponse(http.StatusOK, userJson, this.Ctx)
}

// 更新用户信息
func (this *UserInfoController) UpdateUserInfoByID() {
	userID := this.Ctx.Input.Param(":id")
	if userID == "" {
		response.SetErrorResponse(this.Ctx, errcode.E.User.ParamError.WithMessage("get userid faild!"))
		return
	}
	var updateUserData auths.Users
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &updateUserData); err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.User.ParamError.WithMessage("parse form faild!"))
		return
	}
	if err := models.UpdateUserInfoByID(userID, &updateUserData); err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.User.UpdateFailed.WithMessage("update userinfo from db faild!"))
		return
	}
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}

// 删除用户
func (this *UserInfoController) DeleteUserInfoByID() {
	userID := this.Ctx.Input.Param(":id")
	if userID == "" {
		response.SetErrorResponse(this.Ctx, errcode.E.User.ParamError.WithMessage("get userid faild!"))
		return
	}
	// 从数据库中删除用户
	err := models.DeleteUserInfoByID(userID)
	if err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.User.DeleteFailed.WithMessage("delete userinfo from db faild!"))
		return
	}
	// 响应
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}
