package version

import (
	"encoding/json"
	"net/http"

	"k8s.io/klog/v2"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/versions"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type VersionController struct {
	beego.Controller
}

func (this *VersionController) GetVersion() {
	info := versions.Info()
	body, err := json.Marshal(info)
	if err != nil {
		klog.Error("json.Unmarshal is err:", err)
	}
	response.HttpResponse(http.StatusOK, body, this.Ctx)
}
