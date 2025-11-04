package service

import (
	"encoding/json"
	"net/http"

	"github.com/astaxie/beego"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/models"
	"github.com/rouroumaibing/go-devops/pkg/utils/errcode"
	"github.com/rouroumaibing/go-devops/pkg/utils/response"
)

type PipelineController struct {
	beego.Controller
}

func (p *PipelineController) GetPipelineById() {
	pipelineID := p.Ctx.Input.Param(":id")
	if pipelineID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.ParamError.WithMessage("invalid pipeline ID"))
		return
	}

	pipeline, err := models.GetPipelineById(pipelineID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.NotFound.WithMessage("pipeline not found: "+err.Error()))
		return
	}

	pipelineJson, err := json.Marshal(pipeline)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal pipeline failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, pipelineJson, p.Ctx)
}

func (p *PipelineController) CreatePipeline() {
	var pipeline service.Pipeline
	if err := json.Unmarshal(p.Ctx.Input.RequestBody, &pipeline); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.ParamError.WithMessage("invalid pipeline param: "+err.Error()))
		return
	}

	if err := models.CreatePipeline(&pipeline, pipeline.PipelineStages); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.CreateFailed.WithMessage("create pipeline failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *PipelineController) UpdatePipelineById() {
	PipelineID := p.Ctx.Input.Param(":id")
	if PipelineID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.IDInvalid.WithMessage("invalid pipeline ID"))
		return
	}
	_, err := models.GetPipelineById(PipelineID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.NotFound.WithMessage("pipeline not found: "+err.Error()))
		return
	}
	var updatePipeline service.Pipeline
	if err := json.Unmarshal(p.Ctx.Input.RequestBody, &updatePipeline); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.ParamError.WithMessage("invalid pipeline param: "+err.Error()))
		return
	}

	if err := models.UpdatePipelineById(&updatePipeline, updatePipeline.PipelineStages); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.UpdateFailed.WithMessage("update pipeline failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *PipelineController) DeletePipelineById() {
	PipelineID := p.Ctx.Input.Param(":id")
	if PipelineID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.IDInvalid.WithMessage("invalid pipeline ID"))
		return
	}
	_, err := models.GetPipelineById(PipelineID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.NotFound.WithMessage("pipeline not found: "+err.Error()))
		return
	}

	if err = models.DeletePipelineById(PipelineID); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.DeleteFailed.WithMessage("delete pipeline failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *PipelineController) RunPipelineJob() {
	PipelineID := p.Ctx.Input.Param(":id")
	if PipelineID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.IDInvalid.WithMessage("invalid pipeline ID"))
		return
	}

	if err := models.RunPipelineJob(PipelineID); err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.CreateFailed.WithMessage("create pipeline job failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, nil, p.Ctx)
}

func (p *PipelineController) GetPipelineJob() {
	PipelineID := p.Ctx.Input.Param(":id")
	if PipelineID == "" {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.IDInvalid.WithMessage("invalid pipeline ID"))
		return
	}

	pipelineJob, err := models.GetPipelineJob(PipelineID)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Pipeline.NotFound.WithMessage("pipeline job not found: "+err.Error()))
		return
	}
	pipelineJobJson, err := json.Marshal(pipelineJob)
	if err != nil {
		response.SetErrorResponse(p.Ctx, errcode.E.Base.MarshalFailed.WithMessage("marshal pipeline job failed: "+err.Error()))
		return
	}
	response.HttpResponse(http.StatusOK, pipelineJobJson, p.Ctx)
}
