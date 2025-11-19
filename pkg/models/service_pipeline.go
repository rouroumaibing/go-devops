package models

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/google/uuid"
	"github.com/rouroumaibing/go-devops/pkg/apis/service"
	"k8s.io/klog/v2"
)

func init() {
	orm.RegisterModel(new(service.Pipeline), new(service.PipelineStage), new(service.PipelineStageGroupJobs), new(service.PipelineStageJobs))
}

func GetPipelineById(pipelineID string) (*service.Pipeline, error) {
	dao := orm.NewOrm()

	err := dao.Begin()
	if err != nil {
		return nil, fmt.Errorf("开启事务失败: %w", err)
	}
	// 只读无需回滚
	defer func() {
		if err := dao.Rollback(); err != nil {
			klog.Errorf("事务回滚失败: %v", err)
		}
	}()

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
	defer func() {
		if r := recover(); r != nil {
			rollbackErr := dao.Rollback()
			if rollbackErr != nil {
				log.Printf("事务回滚失败: %v", rollbackErr)
			}
		}
	}()

	pipeline_sql := `INSERT INTO pipeline(id, name, pipeline_group, component_id, service_tree, created_at, updated_at) VALUES(?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`
	_, err = dao.Raw(pipeline_sql, pipelineID, p.Name, p.PipelineGroup, p.ComponentId, p.ServiceTree).Exec()
	if err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			return fmt.Errorf("创建流水线失败: %v, 回滚错误: %v", err, rollbackErr)
		}
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
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				return fmt.Errorf("创建流水线任务失败: %v, 回滚错误: %v", err, rollbackErr)
			}
			return fmt.Errorf("创建流水线任务失败: %v", err)
		}
	}

	if err := dao.Commit(); err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			return fmt.Errorf("提交事务失败: %v, 回滚错误: %v", err, rollbackErr)
		}
		return fmt.Errorf("提交事务失败: %v", err)
	}

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
	defer func() {
		if r := recover(); r != nil {
			rollbackErr := dao.Rollback()
			if rollbackErr != nil {
				log.Printf("事务回滚失败: %v", rollbackErr)
			}
		}
	}()

	if len(pipeline_fields) != 0 {
		pipeline_sql += ", " + strings.Join(pipeline_fields, ", ") + " WHERE id = ?"
		pipeline_params = append(pipeline_params, p.Id)

		_, err = dao.Raw(pipeline_sql, pipeline_params...).Exec()
		if err != nil {
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				return fmt.Errorf("提交事务失败: %v, 回滚错误: %v", err, rollbackErr)
			}
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
					if rollbackErr := dao.Rollback(); rollbackErr != nil {
						return fmt.Errorf("提交事务失败: %v, 回滚错误: %v", err, rollbackErr)
					}
					return fmt.Errorf("更新流水线阶段失败: %v", err)
				}
			}
		}
	}

	// 提交事务
	if err := dao.Commit(); err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			return fmt.Errorf("提交事务失败: %v, 回滚错误: %v", err, rollbackErr)
		}
		return fmt.Errorf("事务提交失败: %v", err)
	}

	return nil
}

func DeletePipelineById(pipelineID string) error {
	dao := orm.NewOrm()
	err := dao.Begin()
	if err != nil {
		return fmt.Errorf("开启事务失败: %v", err)
	}

	defer func() {
		if rollbackErr := recover(); rollbackErr != nil {
			_ = dao.Rollback()
			klog.Errorf("事务回滚失败: %v", rollbackErr)
			panic(fmt.Errorf("panic in DeletePipelineById: %v\n%s", rollbackErr, debug.Stack()))
		}
	}()

	var exists int
	err = dao.Raw(`SELECT 1 FROM pipeline WHERE id = ? FOR UPDATE`, pipelineID).QueryRow(&exists)
	if err != nil {
		if rollbackErr := dao.Rollback(); rollbackErr != nil {
			klog.Errorf("事务回滚失败: %v", rollbackErr)
		}
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
			if rollbackErr := dao.Rollback(); rollbackErr != nil {
				return fmt.Errorf("删除%s失败: %w, 回滚错误: %v", sqlCommands.name, err, rollbackErr)
			}
			return fmt.Errorf("删除%s失败: %w", sqlCommands.name, err)
		}
	}

	return dao.Commit()
}

// CheckIsLeader 检查当前节点是否为主节点
func CheckIsLeader() (bool, error) {
	// 查询数据库中的leader_election表，获取当前集群的主节点信息
	dao := orm.NewOrm()
	var leaderInfo struct {
		LeaderID string
	}

	// 这里假设只有一个集群，或者可以通过环境变量获取集群ID
	err := dao.Raw("SELECT leader_id FROM leader_election LIMIT 1").QueryRow(&leaderInfo)
	if err != nil {
		if err == orm.ErrNoRows {
			return false, nil // 没有主节点信息
		}
		return false, fmt.Errorf("查询主节点信息失败: %v", err)
	}

	// 获取当前节点ID，这里假设节点ID存储在环境变量中
	currentNodeID := os.Getenv("NODE_ID")
	if currentNodeID == "" {
		return false, fmt.Errorf("未设置NODE_ID环境变量")
	}

	// 检查当前节点是否为主节点
	return currentNodeID == leaderInfo.LeaderID, nil
}

func RunPipelineJob(pipelineId string) error {
	// 调用CheckIsLeader检查当前节点是否为主节点
	isLeader, err := CheckIsLeader()
	if err != nil {
		return fmt.Errorf("检查主节点状态失败: %v", err)
	}
	if !isLeader {
		return fmt.Errorf("当前不是主节点，无法生成任务")
	}

	// 事务处理
	o := orm.NewOrm()
	err = o.Begin()
	if err != nil {
		return fmt.Errorf("开始事务失败: %v", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := o.Rollback(); rollbackErr != nil {
				fmt.Printf("回滚事务失败: %v\n", rollbackErr)
			}
		} else {
			if commitErr := o.Commit(); commitErr != nil {
				fmt.Printf("提交事务失败: %v\n", commitErr)
			}
		}
	}()

	// 根据数据库已有的流水线详细信息，生成任务
	// 1.根据pipelineID查询PipelineStage表获取全部的任务
	// 2.根据PipelineStage表中的任务，生成PipelineStageGroupJobs、PipelineStageJobs
	// 2.1根据StageGroupId分组，插入任务到PipelineStageGroupJobs
	// 2.2根据StageId分组，插入任务到PipelineStageJobs
	var pipelineStages []service.PipelineStage
	_, err = orm.NewOrm().QueryTable(&service.PipelineStage{}).Filter("PipelineId", pipelineId).All(&pipelineStages)
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

		_, err = o.Insert(groupJob)
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

			_, err = o.Insert(stageJob)
			if err != nil {
				return fmt.Errorf("插入阶段任务失败: %w", err)
			}

			stageJobIds = append(stageJobIds, stageJob.Id)
		}

		groupJob.StageJobIds = strings.Join(stageJobIds, ",")
		_, err = o.Update(groupJob)
		if err != nil {
			return fmt.Errorf("更新阶段组任务ID列表失败: %w", err)
		}
	}
	return nil
}

func GetPipelineJob(pipelineId string) (service.PipelineStageGroupJobs, error) {
	// 查询PipelineStageGroupJobs，返回切片
	var pipelinejobs service.PipelineStageGroupJobs
	_, err := orm.NewOrm().QueryTable(&service.PipelineStageGroupJobs{}).Filter("PipelineId", pipelineId).All(&pipelinejobs)
	if err != nil {
		return service.PipelineStageGroupJobs{}, fmt.Errorf("查询流水线阶段组任务失败: %w", err)
	}

	var stageJobs []service.PipelineStageJobs
	_, err = orm.NewOrm().QueryTable(&service.PipelineStageJobs{}).Filter("StageGroupJobId", pipelinejobs.Id).All(&stageJobs)
	if err != nil {
		return service.PipelineStageGroupJobs{}, fmt.Errorf("查询流水线阶段任务失败: %w", err)
	}
	pipelinejobs.PipelineStageJobs = append(pipelinejobs.PipelineStageJobs, stageJobs...)

	return pipelinejobs, nil
}
