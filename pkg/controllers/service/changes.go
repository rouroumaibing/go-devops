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

type ChangeController struct {
	beego.Controller
}

func (c *ChangeController) GetChangeById() {
	changeID := c.Ctx.Input.Param(":id")
	if changeID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.ParamError.WithMessage("parameter error: "+fmt.Errorf("change ID error").Error()))
		return
	}

	change, err := models.GetChangeById(changeID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.NotFound.WithMessage("get change failed: "+err.Error()))
		return
	}
	changesJson, err := json.Marshal(change)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.NotFound.WithMessage("get change failed: "+err.Error()))
		return
	}

	response.HttpResponse(http.StatusOK, changesJson, c.Ctx)
}

func (c *ChangeController) CreateChange() {
	var change service.ChangeLog

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &change); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}
	if err := models.CreateChange(&change); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.CreateFailed.WithMessage("create change failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}

func (c *ChangeController) UpdateChangeById() {
	changeID := c.Ctx.Input.Param(":id")
	if changeID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.UpdateFailed.WithMessage("update change failed: "+fmt.Errorf("change ID error").Error()))
		return
	}
	_, err := models.GetChangeById(changeID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.UpdateFailed.WithMessage("update change failed: "+err.Error()))
		return
	}
	var updateChange service.ChangeLog

	if err = json.Unmarshal(c.Ctx.Input.RequestBody, &updateChange); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.ParamError.WithMessage("parameter error: "+err.Error()))
		return
	}

	if err = models.UpdateChangeById(&updateChange); err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.UpdateFailed.WithMessage("update change failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}

func (c *ChangeController) DeleteChangeById() {
	changeID := c.Ctx.Input.Param(":id")
	if changeID == "" {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.DeleteFailed.WithMessage("delete change failed: "+fmt.Errorf("change ID error").Error()))
		return
	}

	_, err := models.GetChangeById(changeID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.DeleteFailed.WithMessage("delete change failed: "+err.Error()))
		return
	}

	err = models.DeleteChangeById(changeID)
	if err != nil {
		response.SetErrorResponse(c.Ctx, errcode.E.Change.DeleteFailed.WithMessage("delete change failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, c.Ctx)
}
