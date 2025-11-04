package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type ComponentController struct {
	beego.Controller
}

func (c *ComponentController) GetComponent() {
	components, err := models.GetComponent()
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component failed: "+err.Error()))
		return
	}
	componentsJson, err := json.Marshal(components)

	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal component data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, componentsJson, c.Ctx)
}
func (c *ComponentController) GetComponentById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	component, err := models.GetComponentById(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.NotFound.WithMessage("component not found: "+err.Error()))
		return
	}
	componentsJson, err := json.Marshal(component)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal component data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, componentsJson, c.Ctx)
}
func (c *ComponentController) CreateComponent() {
	var component service.Component
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &component); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}
	componentId, err := models.CreateComponent(&component)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.CreateFailed.WithMessage("create component failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, []byte(componentId), c.Ctx)
}

func (c *ComponentController) UpdateComponentById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.UpdateFailed.WithMessage("update component failed: "+fmt.Errorf("component ID error").Error()))
		return
	}
	_, err := models.GetComponentById(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.UpdateFailed.WithMessage("update component failed: "+err.Error()))
		return
	}
	var updateComponent service.Component
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &updateComponent); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}
	if err := models.UpdateComponentById(&updateComponent); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.UpdateFailed.WithMessage("update component failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}
func (c *ComponentController) DeleteComponentById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.DeleteFailed.WithMessage("delete component failed: "+fmt.Errorf("component ID error").Error()))
		return
	}
	_, err := models.GetComponentById(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.DeleteFailed.WithMessage("delete component failed: "+err.Error()))
		return
	}

	if err := models.DeleteComponentById(componentID); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.DeleteFailed.WithMessage("delete component failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}

func (c *ComponentController) GetComponentPipelines() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	pipelines, err := models.GetComponentPipelines(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component pipelines failed: "+err.Error()))
		return
	}
	pipelinesJson, err := json.Marshal(pipelines)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal pipelines data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, pipelinesJson, c.Ctx)
}

func (c *ComponentController) GetComponentEnvs() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	envs, err := models.GetComponentEnvs(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component envs failed: "+err.Error()))
		return
	}
	envsJson, err := json.Marshal(envs)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal envs data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, envsJson, c.Ctx)
}

func (c *ComponentController) GetComponentProducts() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	products, err := models.GetComponentProducts(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component products failed: "+err.Error()))
		return
	}
	productsJson, err := json.Marshal(products)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal products data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, productsJson, c.Ctx)
}

func (c *ComponentController) GetComponentChanges() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	changes, err := models.GetComponentChanges(componentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component changes failed: "+err.Error()))
		return
	}
	changesJson, err := json.Marshal(changes)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal changes data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, changesJson, c.Ctx)
}

func (c *ComponentController) CreateComponentParameter() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	var componentParameter service.ComponentParameter
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &componentParameter); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}

	if err := models.CreateComponentParameter(&componentParameter); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.CreateFailed.WithMessage("create component parameter failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}
func (c *ComponentController) GetComponentParameterById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	environmentID := c.Ctx.Input.Param(":environment_id")
	if environmentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("environment ID error").Error()))
		return
	}
	componentParameter, err := models.GetComponentParameterById(componentID, environmentID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.QueryFailed.WithMessage("get component parameter failed: "+err.Error()))
		return
	}
	componentParameterJson, err := json.Marshal(componentParameter)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal component parameter data failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, componentParameterJson, c.Ctx)
}

func (c *ComponentController) UpdateComponentParameterById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	environmentID := c.Ctx.Input.Param(":environment_id")
	if environmentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("environment ID error").Error()))
		return
	}
	var componentParameter service.ComponentParameter
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &componentParameter); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}
	if err := models.UpdateComponentParameterById(componentID, environmentID, &componentParameter); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.UpdateFailed.WithMessage("update component parameter failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}

func (c *ComponentController) DeleteComponentParameterById() {
	componentID := c.Ctx.Input.Param(":id")
	if componentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("component ID error").Error()))
		return
	}
	environmentID := c.Ctx.Input.Param(":environment_id")
	if environmentID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.ParamError.WithMessage("parameter error: "+fmt.Errorf("environment ID error").Error()))
		return
	}
	if err := models.DeleteComponentParameterById(componentID, environmentID); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Component.DeleteFailed.WithMessage("delete component parameter failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}
