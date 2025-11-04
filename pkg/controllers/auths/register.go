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

type RegisterController struct {
	beego.Controller
}

func (this *RegisterController) Register() {

	var form *auths.Users
	if err := json.Unmarshal(this.Ctx.Input.RequestBody, &form); err != nil {
		klog.Error("json.Unmarshal is err:", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Internal.Internal.WithMessage("request body error!"))
		return
	}

	err := models.RegisterUserWithPWD(form)
	if err != nil {
		klog.Error("register faild: ", err)
		response.SetErrorResponse(this.Ctx, errcode.E.Register.RegisterFaild.WithMessage("register faild!: "+err.Error()))
	} else {
		response.HttpResponse(http.StatusOK, nil, this.Ctx)
	}
}
