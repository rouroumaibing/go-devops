package servicetree

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/servicetree"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type ServiceTreeController struct {
	beego.Controller
}

func (s *ServiceTreeController) GetServiceTree() {
	serviceTree, err := models.GetServiceTree()
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.QueryFailed.WithMessage("get servicetree failed: "+err.Error()))
		return
	}

	if serviceTree == nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.NotFound.WithMessage("servicetree not found"))
		return
	}

	serviceTreeJson, err := json.Marshal(serviceTree)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.QueryFailed.WithMessage("get servicetree failed: "+err.Error()))
		return
	}

	response.HttpResponse(http.StatusOK, serviceTreeJson, s.Ctx)
}

func (s *ServiceTreeController) GetServiceTreeById() {
	serviceTreeID := s.Ctx.Input.Param(":id")
	if serviceTreeID == "" {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.ParamError.WithMessage("invalid servicetree ID"))
		return
	}

	serviceTree, err := models.GetServiceTreeById(serviceTreeID)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.QueryFailed.WithMessage("get servicetree failed: "+err.Error()))
		return
	}

	if serviceTree == nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.NotFound.WithMessage("servicetree not found"))
		return
	}
	serviceTreeJson, err := json.Marshal(serviceTree)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.QueryFailed.WithMessage("get servicetree failed: "+err.Error()))
		return
	}

	response.HttpResponse(http.StatusOK, serviceTreeJson, s.Ctx)
}

func (s *ServiceTreeController) CreateServiceTree() {
	var serviceTree servicetree.ServiceTree

	if err := json.Unmarshal(s.Ctx.Input.RequestBody, &serviceTree); err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.ParamError.WithMessage("create servicetree failed: "+err.Error()))
		return
	}

	if err := models.CreateServiceTree(&serviceTree); err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.CreateFailed.WithMessage("create servicetree failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, s.Ctx)
}

func (s *ServiceTreeController) UpdateServiceTreeById() {
	serviceTreeID := s.Ctx.Input.Param(":id")
	if serviceTreeID == "" {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.IDInvalid.WithMessage("invalid serviceTree ID"))
		return
	}
	_, err := models.GetServiceTreeById(serviceTreeID)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.UpdateFailed.WithMessage("update servicetree failed: "+err.Error()))
		return
	}

	var updateServiceTree servicetree.ServiceTree
	if err := json.Unmarshal(s.Ctx.Input.RequestBody, &updateServiceTree); err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.ParamError.WithMessage("update servicetree failed: "+err.Error()))
		return
	}

	if err := models.UpdateServiceTreeById(&updateServiceTree); err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.UpdateFailed.WithMessage("update servicetree failed: "+err.Error()))
		return
	}

	response.HttpResponse(http.StatusOK, nil, s.Ctx)
}
func (s *ServiceTreeController) DeleteServiceTreeById() {
	deleteServiceTreeID := s.Ctx.Input.Param(":id")
	if deleteServiceTreeID == "" {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.DeleteFailed.WithMessage("delete servicetree failed: "+fmt.Errorf("serviceTree ID error").Error()))
		return
	}
	_, err := models.GetServiceTreeById(deleteServiceTreeID)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.DeleteFailed.WithMessage("delete servicetree failed: "+err.Error()))
		return
	}

	err = models.DeleteServiceTreeById(deleteServiceTreeID)
	if err != nil {
		response.SetErrorResponse(s.Ctx, errcode.E.Service.DeleteFailed.WithMessage("delete servicetree failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, s.Ctx)
}
