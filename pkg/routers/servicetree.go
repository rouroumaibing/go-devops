package routers

import (
	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/controllers/servicetree"
)

func init() {
	ServiceTree()
}

// servicetree manage
func ServiceTree() {
	beego.Router("/api/servicetree", &servicetree.ServiceTreeController{}, "get:GetServiceTree")
	beego.Router("/api/servicetree/:id", &servicetree.ServiceTreeController{}, "get:GetServiceTreeById")
	beego.Router("/api/servicetree/:id", &servicetree.ServiceTreeController{}, "put:UpdateServiceTreeById")
	beego.Router("/api/servicetree", &servicetree.ServiceTreeController{}, "post:CreateServiceTree")
	beego.Router("/api/servicetree/:id", &servicetree.ServiceTreeController{}, "delete:DeleteServiceTreeById")
}
