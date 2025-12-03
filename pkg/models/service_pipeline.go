package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"runtime/debug"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
	"github.com/rouroumaibing/go-devops/pkg/workers"
)

func init() {
	orm.RegisterModel(new(service.Pipeline), new(service.PipelineStage), new(service.PipelineStageGroupJobs), new(service.PipelineStageJobs))
}

func GetPipelineById(pipelineID string) (*service.Pipeline, error) {
	dao := orm.NewOrm()

	// 查询pipelines
	pipeline := &service.Pipeline{Id: pipelineID}
	if err := dao.Read(pipeline); err != nil {
		if err == orm.ErrNoRows {
			return nil, fmt.Errorf("流水线不存在, ID: %s", pipelineID)
		}
		return nil, fmt.Errorf("查询流水线失败: %v", err)
	}

	// 查询stages
	var stages []*service.PipelineStage
	if _, err := dao.QueryTable(new(service.PipelineStage)).
		Filter("pipeline_id", pipelineID).
		OrderBy("stage_group_order", "stage_order").
		All(&stages); err != nil {
		return nil, fmt.Errorf("查询流水线阶段失败: %v", err)
	}
	pipeline.PipelineStages = stages

	return pipeline, nil
}

func CreatePipeline(p *service.Pipeline, s []*service.PipelineStage) error {
	if p.ComponentId == "" || p.ServiceTree == "" {
		return fmt.Errorf("组件ID和服务树不能为空")
	}

	lock.Lock()
	defer lock.Unlock()

	pipelineID := uuid.New().String()
	dao := orm.NewOrm()
	err := dao.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}

	// 使用committed标记确保事务只回滚一次
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				log.Printf("CreatePipeline: 事务回滚失败, pipelineID: %s, err: %v", pipelineID, rollbackErr)
			}
			// 如果是panic，重新抛出
			if r := recover(); r != nil {
				log.Printf("CreatePipeline: 事务回滚时捕获到panic, pipelineID: %s, panic: %v", pipelineID, r)
				panic(fmt.Errorf("panic in CreatePipeline: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	pipeline_sql := `INSERT INTO pipeline(id, name, pipeline_group, component_id, service_tree, created_at, updated_at) VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err = dao.Raw(pipeline_sql, pipelineID, p.Name, p.PipelineGroup, p.ComponentId, p.ServiceTree).Exec()
	if err != nil {
		return fmt.Errorf("创建流水线失败: %v", err)
	}

	// stageGroupOrder、stageGroupName前端传入，多个stageGroupName和stageGroupID一一对应
	stageGroupMap := make(map[string]string)

	stage_sql := `INSERT INTO pipeline_stage(id, stage_group_id, stage_group_name, stage_group_order, stage_type, stage_name, parallel, stage_order, job_type,  parameters, pipeline_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	for _, stage := range s {
		stageID := uuid.New().String()
		// 检查group_name是否已存在映射，不存在则生成新groupID
		stageGroupID, exists := stageGroupMap[stage.StageGroupName]
		if !exists {
			stageGroupID = uuid.New().String()
			stageGroupMap[stage.StageGroupName] = stageGroupID
		}
		_, err = dao.Raw(stage_sql, stageID, stageGroupID, stage.StageGroupName, stage.StageGroupOrder, stage.StageType, stage.StageName, stage.Parallel, stage.StageOrder, stage.JobType, stage.Parameters, pipelineID).Exec()
		if err != nil {
			return fmt.Errorf("创建流水线任务失败: %v", err)
		}
	}

	if err := dao.Commit(); err != nil {
		// 提交失败时，事务已经被关闭，不需要再回滚
		// 设置committed为true，防止defer函数尝试回滚
		committed = true
		return fmt.Errorf("提交事务失败: %v", err)
	}

	committed = true
	return nil
}

func UpdatePipelineById(p *service.Pipeline, s []*service.PipelineStage) error {
	pipeline_sql := "UPDATE pipeline SET updated_at = CURRENT_TIMESTAMP"
	stage_sql := "UPDATE pipeline_stage SET updated_at = CURRENT_TIMESTAMP"

	pipeline_params := []interface{}{}
	pipeline_fields := []string{}

	if p.Name != "" {
		pipeline_fields = append(pipeline_fields, "name = ?")
		pipeline_params = append(pipeline_params, p.Name)
	}
	if p.PipelineGroup != "" {
		pipeline_fields = append(pipeline_fields, "pipeline_group = ?")
		pipeline_params = append(pipeline_params, p.PipelineGroup)
	}
	if p.ComponentId != "" {
		pipeline_fields = append(pipeline_fields, "component_id = ?")
		pipeline_params = append(pipeline_params, p.ComponentId)
	}
	if p.ServiceTree != "" {
		pipeline_fields = append(pipeline_fields, "service_tree = ?")
		pipeline_params = append(pipeline_params, p.ServiceTree)
	}

	lock.Lock()
	defer lock.Unlock()

	dao := orm.NewOrm()
	err := dao.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}

	// 使用committed标记确保事务只回滚一次
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				log.Printf("UpdatePipelineById: 事务回滚失败, pipelineID: %s, err: %v", p.Id, rollbackErr)
			}
			if r := recover(); r != nil {
				log.Printf("UpdatePipelineById: 事务回滚时捕获到panic, pipelineID: %s, panic: %v", p.Id, r)
				panic(fmt.Errorf("panic in UpdatePipelineById: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	if len(pipeline_fields) != 0 {
		pipeline_sql += ", " + strings.Join(pipeline_fields, ", ") + " WHERE id = ?"
		pipeline_params = append(pipeline_params, p.Id)

		_, err = dao.Raw(pipeline_sql, pipeline_params...).Exec()
		if err != nil {
			return fmt.Errorf("更新流水线失败: %v", err)
		}
	}

	// 更新pipeline_stage表
	if len(s) > 0 {
		for _, stage := range s {
			stageParams := []interface{}{}
			stageFields := []string{}

			if stage.StageGroupOrder != 0 {
				stageFields = append(stageFields, "stage_group_order = ?")
				stageParams = append(stageParams, stage.StageGroupOrder)
			}
			if stage.StageOrder != 0 {
				stageFields = append(stageFields, "stage_order = ?")
				stageParams = append(stageParams, stage.StageOrder)
			}
			if stage.StageType != "" {
				stageFields = append(stageFields, "stage_type = ?")
				stageParams = append(stageParams, stage.StageType)
			}
			if stage.StageName != "" {
				stageFields = append(stageFields, "stage_name = ?")
				stageParams = append(stageParams, stage.StageName)
			}
			if stage.JobType != "" {
				stageFields = append(stageFields, "job_type = ?")
				stageParams = append(stageParams, stage.JobType)
			}
			if stage.Parallel {
				stageFields = append(stageFields, "parallel = ?")
				stageParams = append(stageParams, stage.Parallel)
			}

			if len(stageFields) != 0 {
				stage_sql += ", " + strings.Join(stageFields, ", ") + " WHERE pipeline_id = ? AND id = ?"
				stageParams = append(stageParams, p.Id, stage.Id)
				_, err = dao.Raw(stage_sql, stageParams...).Exec()
				if err != nil {
					return fmt.Errorf("更新流水线阶段失败: %v", err)
				}
			}
		}
	}

	// 提交事务
	if err := dao.Commit(); err != nil {
		// 提交失败时，事务已经被关闭，不需要再回滚
		// 设置committed为true，防止defer函数尝试回滚
		committed = true
		return fmt.Errorf("事务提交失败: %v", err)
	}

	committed = true
	return nil
}

func DeletePipelineById(pipelineID string) error {
	dao := orm.NewOrm()
	err := dao.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}

	// 使用committed标记确保事务只回滚一次
	committed := false
	defer func() {
		if !committed {
			// 统一回滚，不关心具体原因
			log.Printf("DeletePipelineById: 事务回滚, pipelineID: %s", pipelineID)
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				log.Printf("DeletePipelineById: 事务回滚失败, pipelineID: %s, err: %v", pipelineID, rollbackErr)
				// 忽略tx closed错误，因为这通常是由于事务已经被关闭导致的
				if !strings.Contains(rollbackErr.Error(), "tx closed") {
					log.Printf("DeletePipelineById: 事务回滚失败(非tx closed), pipelineID: %s, err: %v", pipelineID, rollbackErr)
				}
			}

			if r := recover(); r != nil {
				log.Printf("DeletePipelineById: 事务回滚时捕获到panic, pipelineID: %s, panic: %v", pipelineID, r)
				panic(fmt.Errorf("panic in DeletePipelineById: %v\n%s", r, debug.Stack()))
			}
		} else {
			log.Printf("DeletePipelineById: 事务已提交, pipelineID: %s", pipelineID)
		}
	}()

	var exists int
	err = dao.Raw(`SELECT 1 FROM pipeline WHERE id = ? FOR UPDATE`, pipelineID).QueryRow(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("流水线不存在: %w", err)
		}
		return fmt.Errorf("查询流水线失败: %w", err)
	}

	for _, sqlCommands := range []struct{ sql, name string }{
		{`DELETE FROM pipeline_stage WHERE pipeline_id = ?`, "阶段"},
		{`DELETE FROM pipeline WHERE id = ?`, "流水线"},
	} {
		if _, err = dao.Raw(sqlCommands.sql, pipelineID).Exec(); err != nil {
			return fmt.Errorf("删除%s失败: %w", sqlCommands.name, err)
		}
	}

	if err := dao.Commit(); err != nil {
		// 提交失败时，事务已经被关闭，不需要再回滚
		// 设置committed为true，防止defer函数尝试回滚
		committed = true
		return fmt.Errorf("提交事务失败: %w", err)
	}

	committed = true
	return nil
}

func RunPipelineJob(pipelineId string) error {
	localNodeIP, err := system.GetLocalIP()
	if err != nil {
		return fmt.Errorf("获取本地IP失败: %v", err)
	}

	// 调用IsLeader检查当前节点是否为主节点
	if isLeader := workers.IsLeader(localNodeIP); !isLeader {
		return fmt.Errorf("当前不是主节点，无法生成任务")
	}

	// 事务处理
	dao := orm.NewOrm()
	err = dao.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}

	// 使用committed标记确保事务只回滚一次
	committed := false
	defer func() {
		if !committed {
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				log.Printf("RunPipelineJob: 事务回滚失败, pipelineID: %s, err: %v", pipelineId, rollbackErr)
			}
			if r := recover(); r != nil {
				log.Printf("RunPipelineJob: 事务回滚时捕获到panic, pipelineID: %s, panic: %v", pipelineId, r)
				panic(fmt.Errorf("panic in RunPipelineJob: %v\n%s", r, debug.Stack()))
			}
		}
	}()

	// 根据数据库已有的流水线详细信息，生成任务
	// 1.根据pipelineID查询PipelineStage表获取全部的任务
	// 2.根据PipelineStage表中的任务，生成PipelineStageGroupJobs、PipelineStageJobs
	// 2.1根据StageGroupId分组，插入任务到PipelineStageGroupJobs
	// 2.2根据StageId分组，插入任务到PipelineStageJobs
	var pipelineStages []service.PipelineStage
	_, err = dao.QueryTable(&service.PipelineStage{}).Filter("pipeline_id", pipelineId).All(&pipelineStages)
	if err != nil {
		return fmt.Errorf("查询流水线阶段失败: %w", err)
	}
	// 检查是否有流水线阶段
	if len(pipelineStages) == 0 {
		return fmt.Errorf("流水线没有配置任何阶段")
	}

	// 按StageGroupId分组
	stageGroupMap := make(map[string][]service.PipelineStage)
	for _, stage := range pipelineStages {
		stageGroupMap[stage.StageGroupId] = append(stageGroupMap[stage.StageGroupId], stage)
	}

	// 遍历每个阶段组，创建PipelineStageGroupJobs和PipelineStageJobs
	for groupId, stages := range stageGroupMap {
		if len(stages) == 0 {
			continue
		}
		// 同组的StageGroupName、Parallel、StageGroupOrder是一样的
		groupJob := &service.PipelineStageGroupJobs{
			Id:              uuid.New().String(),
			Name:            stages[0].StageGroupName,
			PipelineId:      pipelineId,
			Parallel:        stages[0].Parallel,
			StageGroupId:    groupId,
			StageGroupOrder: stages[0].StageGroupOrder,
			Status:          service.PipelineJobStatusInit,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		_, err = dao.Insert(groupJob)
		if err != nil {
			return fmt.Errorf("插入阶段组任务失败: %w", err)
		}

		// 创建PipelineStageJobs记录并收集阶段作业ID
		var stageJobIds []string
		for _, stage := range stages {
			pipelineJobParams := stage.Parameters
			if pipelineJobParams == "" {
				pipelineJobParams = "{}"
			}

			stageJob := &service.PipelineStageJobs{
				Id:              uuid.New().String(),
				Name:            stage.StageName,
				StageOrder:      stage.StageOrder,
				PipelineId:      pipelineId,
				StageGroupJobId: groupJob.Id,
				Parameters:      pipelineJobParams,
				Status:          service.PipelineJobStatusInit,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			_, err = dao.Insert(stageJob)
			if err != nil {
				return fmt.Errorf("插入阶段任务失败: %w", err)
			}

			stageJobIds = append(stageJobIds, stageJob.Id)
		}

		groupJob.StageJobIds = strings.Join(stageJobIds, ",")
		_, err = dao.Update(groupJob)
		if err != nil {
			return fmt.Errorf("更新阶段组任务ID列表失败: %w", err)
		}
	}

	// 提交事务
	if err := dao.Commit(); err != nil {
		// 提交失败时，事务已经被关闭，不需要再回滚
		// 设置committed为true，防止defer函数尝试回滚
		committed = true
		return fmt.Errorf("提交事务失败: %w", err)
	}

	committed = true
	return nil
}

func GetPipelineJob(pipelineId string) ([]service.PipelineStageGroupJobs, error) {
	// 查询PipelineStageGroupJobs，返回切片
	var pipelinejobs []service.PipelineStageGroupJobs
	_, err := orm.NewOrm().QueryTable(&service.PipelineStageGroupJobs{}).Filter("pipeline_id", pipelineId).All(&pipelinejobs)
	if err != nil {
		return nil, fmt.Errorf("查询流水线阶段组任务失败: %w", err)
	}

	// 为每个阶段组任务查询对应的阶段任务
	for job := range pipelinejobs {
		var stageJobs []service.PipelineStageJobs
		_, err = orm.NewOrm().QueryTable(&service.PipelineStageJobs{}).Filter("stage_group_job_id", pipelinejobs[job].Id).All(&stageJobs)
		if err != nil {
			return nil, fmt.Errorf("查询流水线阶段任务失败: %w", err)
		}
		pipelinejobs[job].PipelineStageJobs = append(pipelinejobs[job].PipelineStageJobs, stageJobs...)
	}

	return pipelinejobs, nil
}
