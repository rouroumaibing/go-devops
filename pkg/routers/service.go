package routers

import (
	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/controllers/service"
)

func init() {
	Component()
	Pipeline()
	Environment()
	Product()
	Change()
}

// component manage
func Component() {
	beego.Router("/api/component", &service.ComponentController{}, "get:GetComponent")

	beego.Router("/api/component", &service.ComponentController{}, "post:CreateComponent")
	beego.Router("/api/component/:id", &service.ComponentController{}, "get:GetComponentById")
	beego.Router("/api/component/:id", &service.ComponentController{}, "put:UpdateComponentById")
	beego.Router("/api/component/:id", &service.ComponentController{}, "delete:DeleteComponentById")

	beego.Router("/api/component/:id/pipelines", &service.ComponentController{}, "get:GetComponentPipelines")
	beego.Router("/api/component/:id/environments", &service.ComponentController{}, "get:GetComponentEnvs")
	beego.Router("/api/component/:id/products", &service.ComponentController{}, "get:GetComponentProducts")
	beego.Router("/api/component/:id/changes", &service.ComponentController{}, "get:GetComponentChanges")

	beego.Router("/api/component/:id/parameters", &service.ComponentController{}, "post:CreateComponentParameter")
	beego.Router("/api/component/:id/parameters/:environment_id", &service.ComponentController{}, "get:GetComponentParameterById")
	beego.Router("/api/component/:id/parameters/:environment_id", &service.ComponentController{}, "put:UpdateComponentParameterById")
	beego.Router("/api/component/:id/parameters/:environment_id", &service.ComponentController{}, "delete:DeleteComponentParameterById")

}

// pipline manage
func Pipeline() {
	beego.Router("/api/pipeline", &service.PipelineController{}, "post:CreatePipeline")
	beego.Router("/api/pipeline/:id", &service.PipelineController{}, "get:GetPipelineById")
	beego.Router("/api/pipeline/:id", &service.PipelineController{}, "put:UpdatePipelineById")
	beego.Router("/api/pipeline/:id", &service.PipelineController{}, "delete:DeletePipelineById")

	beego.Router("/api/pipeline/:id/jobs", &service.PipelineController{}, "post:RunPipelineJob")
	beego.Router("/api/pipeline/:id/jobs", &service.PipelineController{}, "get:GetPipelineJob")
}

// evironment manage
func Environment() {
	beego.Router("/api/environment", &service.EnvironmentController{}, "post:CreateEnvironment")
	beego.Router("/api/environment/:id", &service.EnvironmentController{}, "get:GetEnvironmentById")
	beego.Router("/api/environment/:id", &service.EnvironmentController{}, "put:UpdateEnvironmentById")
	beego.Router("/api/environment/:id", &service.EnvironmentController{}, "delete:DeleteEnvironmentById")
}

// product manage
func Product() {
	beego.Router("/api/product", &service.ProductController{}, "post:CreateProduct")
	beego.Router("/api/product/:id", &service.ProductController{}, "get:GetProductById")
	beego.Router("/api/product/:id", &service.ProductController{}, "put:UpdateProductById")
	beego.Router("/api/product/:id", &service.ProductController{}, "delete:DeleteProductById")
}

// build manage
func Change() {
	beego.Router("/api/change", &service.ChangeController{}, "post:CreateChange")
	beego.Router("/api/change/:id", &service.ChangeController{}, "get:GetChangeById")
	beego.Router("/api/change/:id", &service.ChangeController{}, "put:UpdateChangeById")
	beego.Router("/api/change/:id", &service.ChangeController{}, "delete:DeleteChangeById")

}
