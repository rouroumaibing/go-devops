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

type EnvironmentController struct {
	beego.Controller
}

func (e *EnvironmentController) GetEnvironmentById() {
	environmentID := e.Ctx.Input.Param(":id")
	if environmentID == "" {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.ParamError.WithMessage("invalid environment ID"))
		return
	}
	environment, err := models.GetEnvironmentById(environmentID)
	if err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.NotFound.WithMessage("environment not found: "+err.Error()))
		return
	}
	environmentJson, err := json.Marshal(environment)
	if err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal environment failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, environmentJson, e.Ctx)
}

func (e *EnvironmentController) CreateEnvironment() {
	var environment service.Environment
	if err := json.Unmarshal(e.Ctx.Input.RequestBody, &environment); err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.ParamError.WithMessage("create environment failed: "+err.Error()))
		return
	}
	if err := models.CreateEnvironment(&environment); err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.CreateFailed.WithMessage("create environment failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, e.Ctx)
}
func (e *EnvironmentController) UpdateEnvironmentById() {
	environmentID := e.Ctx.Input.Param(":id")
	if environmentID == "" {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.UpdateFailed.WithMessage("update environment failed: "+fmt.Errorf("environment ID error").Error()))
		return
	}
	_, err := models.GetEnvironmentById(environmentID)
	if err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.UpdateFailed.WithMessage("update environment failed: "+err.Error()))
		return
	}
	var updateEnvironment service.Environment
	if err := json.Unmarshal(e.Ctx.Input.RequestBody, &updateEnvironment); err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.ParamError.WithMessage("update environment failed: "+err.Error()))
		return
	}
	if err := models.UpdateEnvironmentById(&updateEnvironment); err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.UpdateFailed.WithMessage("update environment failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, e.Ctx)
}
func (e *EnvironmentController) DeleteEnvironmentById() {
	environmentID := e.Ctx.Input.Param(":id")
	if environmentID == "" {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.DeleteFailed.WithMessage("delete environment failed: "+fmt.Errorf("environment ID error").Error()))
		return
	}
	_, err := models.GetEnvironmentById(environmentID)
	if err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.DeleteFailed.WithMessage("delete environment failed: "+err.Error()))
		return
	}

	if err := models.DeleteEnvironmentById(environmentID); err != nil {
		response.SetErrorResponse(e.Ctx, errcode.E.Environment.DeleteFailed.WithMessage("delete environment failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, e.Ctx)
}
