package auths

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"k8s.io/klog/v2"

	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type LoginController struct {
	beego.Controller
}

func (this *LoginController) RefreshToken() {
	refreshToken := this.Ctx.Input.Header("X-Refresh-Token")
	if refreshToken == "" {
		response.SetErrorResponse(this.Ctx, errcode.E.Base.Token.WithMessage("refresh token is empty"))
		return
	}

	userID, err := models.CheckRefreshToken(refreshToken)
	if err != nil {
		klog.Error("parse refresh token failed: ", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Base.Token.WithMessage("invalid refresh token"))
		return
	}

	err = models.DeleteAccessToken(userID)
	if err != nil {
		klog.Error("delete old access token failed: ", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("refresh token failed"))
		return
	}

	accessToken, err := models.RefreshUserToken(userID, refreshToken)
	if err != nil {
		klog.Error("generate new tokens failed: ", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("refresh token failed"))
		return
	}

	tokenResponse := &auths.TokenResponse{
		AccessToken: accessToken,
	}

	tokenResponseJson, err := json.Marshal(tokenResponse)
	if err != nil {
		klog.Error("json.Marshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("refresh token failed"))
		return
	}

	// 返回成功响应
	response.HttpResponse(http.StatusOK, tokenResponseJson, this.Ctx)
	klog.Info("refresh token success for user: ", userID)
}

func (this *LoginController) LoginPWD() {
	var formPwd *auths.Users
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &formPwd); err != nil {
		klog.Error("json.Unmarshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("request body error!"))
		return
	}

	if useruuid, err := models.LoginWithPwd(formPwd); err != nil {
		klog.Info("Login faild!", formPwd.Accountname)
		response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthPWD.WithMessage("login failed!"))
	} else {
		accessToken, refreshToken, err := models.NewToken(useruuid.Id)
		if err != nil {
			klog.Error(formPwd.Accountname, ": Get token failed: ", err)
			response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthPWD.WithMessage("login failed!"))
			return
		}

		tokenResponse := &auths.TokenResponse{
			UserUUID:     useruuid.Id,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		}

		tokenResponseJson, err := json.Marshal(tokenResponse)
		if err != nil {
			klog.Error("json.Marshal is err:", err)
			response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("login failed!"))
			return
		}
		response.HttpResponse(http.StatusOK, tokenResponseJson, this.Ctx)
		klog.Info("Login Success!", formPwd.Accountname)
	}

}

func (this *LoginController) Logout() {
	useruuid := this.Ctx.Input.GetData("useruuid").(string)
	if err := models.DeleteAccessToken(useruuid); err != nil {
		response.SetErrorResponse(this.Ctx, errcode.E.Base.Token.WithMessage("token delete failed!"))
		return
	}
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}

func (this *LoginController) WeChatCheck(signature, timestamp, nonce, echostr string) string {

	// 校验 ，待完善
	return echostr
}

func (this *LoginController) WeChatCallback() {
	var formWechat *auths.WeChatCallbackRequest
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &formWechat); err != nil {
		klog.Error("json.Unmarshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("request body error!"))
		return
	}
	useruuid, accessToken, refreshToken, err := models.WeChatCallback(formWechat.Code)
	if err != nil {
		klog.Error("WeChatCallback is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("WeChatCallback failed!"))
		return
	}
	tokenResponse := &auths.TokenResponse{
		UserUUID:     useruuid,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	tokenResponseJson, err := json.Marshal(tokenResponse)
	if err != nil {
		klog.Error("json.Marshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("WeChatCallback failed!"))
		return
	}

	response.HttpResponse(http.StatusOK, tokenResponseJson, this.Ctx)
}

func (this *LoginController) QQCallback() {
	// 待完成
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}
