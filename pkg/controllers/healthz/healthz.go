package healthz

import (
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type HealthzController struct {
	beego.Controller
}

func (this *HealthzController) HealthCheck() {
	response.HttpResponse(http.StatusOK, nil, this.Ctx)
}
