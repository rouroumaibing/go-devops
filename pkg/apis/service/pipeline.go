package service

import (
	"time"
)

const (
	PipelineJobStatusInit    = "wait"
	PipelineJobStatusRun     = "processing"
	PipelineJobStatusSuccess = "success"
	PipelineJobStatusFailed  = "failed"
)

type Pipeline struct {
	Id             string           `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name           string           `json:"name" orm:"column(name);size(50);null(false);description(流水线名称)"`
	PipelineGroup  string           `json:"pipeline_group" orm:"column(pipeline_group);size(50);null;description(流水线组)"`
	ComponentId    string           `json:"component_id" orm:"column(component_id);null(false);description(关联组件ID)"`
	ServiceTree    string           `json:"service_tree" orm:"column(service_tree);size(100);null;description(服务树路径)"`
	CreatedAt      time.Time        `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt      time.Time        `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
	PipelineStages []*PipelineStage `json:"pipeline_stages" orm:"-"`
}

type PipelineStage struct {
	Id              string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	StageGroupId    string    `json:"stage_group_id" orm:"column(stage_group_id);null(false);description(阶段组ID)"`
	StageGroupName  string    `json:"stage_group_name" orm:"column(stage_group_name);size(50);null(false);description(阶段组名称,如构建/卡点/部署)"`
	StageGroupOrder int       `json:"stage_group_order" orm:"column(stage_group_order);null(false);description(阶段组排序序号,数字越小越靠前)"`
	StageType       string    `json:"stage_type" orm:"column(stage_type);size(50);null(false);description(阶段类型,如build/checkpoint/deploy/test)"`
	StageName       string    `json:"stage_name" orm:"column(stage_name);size(50);null(false);description(阶段名称,如build_job/checkpoint/deploy_env/test_job)"`
	Parallel        bool      `json:"parallel" orm:"column(parallel);null(false);description(是否并行执行)"`
	StageOrder      int       `json:"stage_order" orm:"column(stage_order);null(false);description(阶段排序序号,数字越小越靠前)"`
	JobType         string    `json:"job_type" orm:"column(job_type);size(50);null;description(作业类型,如build/checkpoint/deploy/test)"`
	Parameters      string    `json:"parameters" orm:"column(parameters);size(2000);null(false);description(阶段参数)"`
	PipelineId      string    `json:"pipeline_id" orm:"column(pipeline_id);null(false);description(流水线ID)"`
	CreatedAt       time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt       time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}

type PipelineStageGroupJobs struct {
	Id                string              `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name              string              `json:"name" orm:"column(name);size(50);null(false);description(阶段组名称,如构建/卡点/部署)"`
	PipelineId        string              `json:"pipeline_id" orm:"column(pipeline_id);null(false);description(流水线ID)"`
	Parallel          bool                `json:"parallel" orm:"column(parallel);null(false);description(是否并行执行)"`
	StageGroupId      string              `json:"stage_group_id" orm:"column(stage_group_id);null(false);description(阶段组ID)"`
	StageGroupOrder   int                 `json:"stage_group_order" orm:"column(stage_group_order);null(false);description(阶段组排序序号,数字越小越靠前)"`
	StageJobIds       string              `json:"stage_job_ids" orm:"column(stage_job_ids);size(2000);null(false);description(阶段组下的作业ID列表,逗号分隔)"`
	Status            string              `json:"status" orm:"column(status);size(20);null;description(阶段状态,如wait, processing, success, failed)"`
	CreatedAt         time.Time           `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt         time.Time           `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
	PipelineStageJobs []PipelineStageJobs `json:"pipeline_stage_jobs" orm:"-"`
}

type PipelineStageJobs struct {
	Id              string    `json:"id" orm:"pk;size(36);column(id);type(string);unique;description(主键ID)"`
	Name            string    `json:"name" orm:"column(name);size(50);null(false);description(阶段名称,如build_job/checkpoint/deploy_env/test_job)"`
	StageOrder      int       `json:"stage_order" orm:"column(stage_order);null(false);description(阶段排序序号,数字越小越靠前)"`
	PipelineId      string    `json:"pipeline_id" orm:"column(pipeline_id);null(false);description(流水线ID)"`
	StageGroupJobId string    `json:"stage_group_job_id" orm:"column(stage_group_job_id);null(false);description(阶段组下的作业ID)"`
	Parameters      string    `json:"parameters" orm:"column(parameters);size(2000);null(false);description(阶段参数)"`
	Status          string    `json:"status" orm:"column(status);size(20);null;description(阶段状态,如wait, processing, success, failed)"`
	CreatedAt       time.Time `json:"created_at" orm:"column(created_at);type(datetime);null;description(创建时间)"`
	UpdatedAt       time.Time `json:"updated_at" orm:"column(updated_at);type(datetime);null;description(更新时间)"`
}
