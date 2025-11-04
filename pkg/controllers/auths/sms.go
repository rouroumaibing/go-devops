package auths

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
	"k8s.io/klog/v2"
)

type SMSController struct {
	beego.Controller
}

func (this *SMSController) GetSMS() {
	phone := this.GetString("phone")
	success := models.SendSMS(phone)
	if !success {
		response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthSMS.WithMessage("send sms faild!"))
		return
	}
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}

func (this *SMSController) SMSLogin() {
	var formSMS *auths.SMS
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &formSMS); err != nil {
		klog.Error("json.Unmarshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("request body error!"))
		return
	}

	if ok := models.ValidateCode(formSMS.Phone, formSMS.Code); !ok {
		klog.Info(formSMS.Phone, ": Login failed! Validate Code failed")
		response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthSMS.WithMessage("login failed!"))
	} else {
		useruuid, err := models.FindUserByPhone(formSMS.Phone)
		if err != nil {
			klog.Error(formSMS.Phone, ": Get useruuid failed: ", err)
			response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthSMS.WithMessage("login failed!"))
			return
		}
		accessToken, refreshToken, err := models.NewToken(useruuid.Id)
		if err != nil {
			klog.Error(formSMS.Phone, ": Get token failed: ", err)
			response.SetErrorResponse(this.Ctx, errcode.E.Base.AuthSMS.WithMessage("login failed!"))
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
	}
}
