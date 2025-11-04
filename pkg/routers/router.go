package routers

import (
	"github.com/astaxie/beego"

	"github.com/rouroumaibing/go-devops/pkg/controllers/healthz"
	"github.com/rouroumaibing/go-devops/pkg/controllers/version"
)

func init() {
	Version()
	HealthCheck()
}

func Version() {
	beego.Router("/api/version", &version.VersionController{}, "get:GetVersion")
}

func HealthCheck() {
	beego.Router("/api/healthz", &healthz.HealthzController{}, "get:HealthCheck")
}
