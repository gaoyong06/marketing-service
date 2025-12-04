package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/go-kratos/kratos/v2/log"
	"marketing-service/internal/constants"
)

// TaskTriggerService 任务触发服务
type TaskTriggerService struct {
	tuc         *TaskUseCase
	tcluc       *TaskCompletionLogUseCase
	guc         *RewardGrantUseCase
	ruc         *RewardUseCase
	cuc         *CampaignUseCase
	iruc        *InventoryReservationUseCase
	validator   *ValidatorService
	generator   *GeneratorService
	distributor *DistributorService
	rmqProducer rocketmq.Producer // 直接使用 RocketMQ Producer
	rmqTopic    string           // RocketMQ Topic 名称
	log         *log.Helper
}

// NewTaskTriggerService 创建任务触发服务
func NewTaskTriggerService(
	tuc *TaskUseCase,
	tcluc *TaskCompletionLogUseCase,
	guc *RewardGrantUseCase,
	ruc *RewardUseCase,
	cuc *CampaignUseCase,
	iruc *InventoryReservationUseCase,
	validator *ValidatorService,
	generator *GeneratorService,
	distributor *DistributorService,
	rmqProducer rocketmq.Producer, // 直接使用 RocketMQ Producer，可为 nil
	rmqTopic string,                // RocketMQ Topic 名称，如果为空则使用默认值
	logger log.Logger,
) *TaskTriggerService {
	// 如果 topic 为空，使用默认值
	if rmqTopic == "" {
		rmqTopic = constants.RocketMQTopicTaskCompleted
	}
	return &TaskTriggerService{
		tuc:         tuc,
		tcluc:       tcluc,
		guc:         guc,
		ruc:         ruc,
		cuc:         cuc,
		iruc:        iruc,
		validator:   validator,
		generator:   generator,
		distributor: distributor,
		rmqProducer: rmqProducer,
		rmqTopic:    rmqTopic,
		log:         log.NewHelper(logger),
	}
}

// TriggerEvent 触发事件（处理任务完成）
func (s *TaskTriggerService) TriggerEvent(ctx context.Context, event *TaskEvent) error {
	// 1. 查询活跃任务
	tasks, err := s.tuc.ListActive(ctx, event.TenantID, event.AppID)
	if err != nil {
		s.log.Errorf("failed to list active tasks: %v", err)
		return err
	}

	// 2. 遍历任务，检查触发条件
	for _, task := range tasks {
		// 检查触发条件
		if !s.matchTrigger(task, event) {
			continue
		}

		// 检查完成条件
		completed, progressData, err := s.checkCondition(ctx, task, event)
		if err != nil {
			s.log.Errorf("failed to check condition for task %s: %v", task.ID, err)
			continue
		}

		if !completed {
			continue
		}

		// 检查任务完成次数限制
		count, err := s.tcluc.CountByTaskAndUser(ctx, task.ID, event.UserID)
		if err != nil {
			s.log.Errorf("failed to count task completions: %v", err)
			continue
		}

		if task.MaxCount > 0 && count >= int64(task.MaxCount) {
			s.log.Infof("task %s max count reached for user %d", task.ID, event.UserID)
			continue
		}

		// 3. 记录任务完成日志
		completionLog := &TaskCompletionLog{
			TaskID:       task.ID,
			TaskName:     task.Name,
			CampaignID:   event.CampaignID,
			CampaignName: event.CampaignName,
			UserID:       event.UserID,
			TenantID:     event.TenantID,
			AppID:        event.AppID,
			ProgressData: progressData,
			TriggerEvent: event.EventType,
			CompletedAt:  time.Now(),
		}

		if _, err := s.tcluc.Create(ctx, completionLog); err != nil {
			s.log.Errorf("failed to create completion log: %v", err)
			continue
		}

		// 4. 如果任务关联了奖励，则发放奖励
		if task.RewardID != "" {
			if err := s.issueReward(ctx, task, event, completionLog); err != nil {
				s.log.Errorf("failed to issue reward: %v", err)
				// 不返回错误，记录日志即可
			}
		}

		// 5. 发布任务完成事件到 RocketMQ（异步处理）
		if s.rmqProducer != nil {
			eventMessage := &TaskEventMessage{
				EventType:    event.EventType,
				UserID:       event.UserID,
				TenantID:     event.TenantID,
				AppID:        event.AppID,
				CampaignID:   event.CampaignID,
				CampaignName: event.CampaignName,
				EventData:    event.EventData,
				Timestamp:    event.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			}

			eventJSON, err := json.Marshal(eventMessage)
			if err != nil {
				s.log.Errorf("failed to marshal task completion event: %v, task=%s, user=%d", err, task.ID, event.UserID)
			} else {
				msg := primitive.NewMessage(s.rmqTopic, eventJSON)
				result, err := s.rmqProducer.SendSync(ctx, msg)
				if err != nil {
					// 连接失败或发送失败时记录错误日志
					s.log.Errorf("failed to publish task completion event to RocketMQ: %v, task=%s, user=%d, event_type=%s", err, task.ID, event.UserID, event.EventType)
				} else {
					s.log.Infof("published task completion event to RocketMQ successfully: task=%s, user=%d, msg_id=%s", task.ID, event.UserID, result.MsgID)
				}
			}
		} else {
			// RocketMQ 未配置时，记录调试日志（可选）
			s.log.Debugf("RocketMQ producer is not available, skipping event publish: task=%s, user=%d", task.ID, event.UserID)
		}
	}

	return nil
}

// matchTrigger 匹配触发条件
func (s *TaskTriggerService) matchTrigger(task *Task, event *TaskEvent) bool {
	if task.TriggerConfig == "" {
		return false
	}

	var triggerConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.TriggerConfig), &triggerConfig); err != nil {
		s.log.Warnf("failed to unmarshal trigger config: %v", err)
		return false
	}

	// 检查事件类型
	if eventType, ok := triggerConfig["event"].(string); ok {
		if eventType != event.EventType {
			return false
		}
	}

	// 可以添加更多触发条件检查
	return true
}

// checkCondition 检查完成条件
func (s *TaskTriggerService) checkCondition(_ context.Context, task *Task, event *TaskEvent) (bool, string, error) {
	if task.ConditionConfig == "" {
		return false, "", nil
	}

	var conditionConfig map[string]interface{}
	if err := json.Unmarshal([]byte(task.ConditionConfig), &conditionConfig); err != nil {
		s.log.Warnf("failed to unmarshal condition config for task %s: %v", task.ID, err)
		return false, "", err
	}

	// 简单的条件检查逻辑
	// 实际应该根据不同的任务类型实现不同的检查逻辑
	conditionType, _ := conditionConfig["type"].(string)
	operator, _ := conditionConfig["operator"].(string)
	value, _ := conditionConfig["value"].(float64)

	// 从事件数据中获取进度值
	progressValue := s.getProgressValue(event, conditionType)

	// 添加调试日志
	s.log.Debugf("checkCondition: task=%s, conditionType=%s, operator=%s, value=%f, progressValue=%f, eventData=%v",
		task.ID, conditionType, operator, value, progressValue, event.EventData)

	// 检查条件
	completed := false
	switch operator {
	case ">=":
		completed = progressValue >= value
	case ">":
		completed = progressValue > value
	case "==":
		completed = progressValue == value
	case "<=":
		completed = progressValue <= value
	case "<":
		completed = progressValue < value
	}

	s.log.Debugf("checkCondition result: task=%s, completed=%v", task.ID, completed)

	// 构建进度数据
	progressData := map[string]interface{}{
		"type":   conditionType,
		"value":  progressValue,
		"target": value,
	}
	progressDataJSON, _ := json.Marshal(progressData)

	return completed, string(progressDataJSON), nil
}

// getProgressValue 从事件中获取进度值
func (s *TaskTriggerService) getProgressValue(event *TaskEvent, conditionType string) float64 {
	// 根据条件类型从事件数据中获取值
	// 这里简化处理，实际应该根据不同的条件类型实现
	if event.EventData != nil {
		val, ok := event.EventData[conditionType]
		if !ok {
			s.log.Debugf("getProgressValue: conditionType %s not found in eventData: %v", conditionType, event.EventData)
			return 0
		}

		// 支持多种类型：float64, int, string
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			// 尝试将字符串转换为数字
			var f float64
			if _, err := fmt.Sscanf(v, "%f", &f); err == nil {
				s.log.Debugf("getProgressValue: converted string %s to float64 %f", v, f)
				return f
			}
			s.log.Warnf("getProgressValue: failed to convert string %s to float64", v)
			return 0
		default:
			s.log.Warnf("getProgressValue: unsupported type %T for value %v", v, v)
			return 0
		}
	}
	return 0
}

// issueReward 发放奖励（完整流程：校验 -> 库存预占 -> 生成 -> 发放）
func (s *TaskTriggerService) issueReward(ctx context.Context, task *Task, event *TaskEvent, log *TaskCompletionLog) error {
	// 1. 获取奖励模板
	reward, err := s.ruc.Get(ctx, task.RewardID)
	if err != nil {
		return err
	}
	if reward == nil {
		return nil
	}

	// 2. 获取活动信息（用于校验）
	var campaign *Campaign
	if event.CampaignID != "" {
		campaign, _ = s.cuc.Get(ctx, event.CampaignID)
	}

	// 3. 校验阶段
	validatorConfig, err := ParseValidatorConfig(reward.ValidatorConfig)
	if err != nil {
		s.log.Warnf("failed to parse validator config: %v", err)
	} else if validatorConfig != nil {
		validationReq := &ValidationRequest{
			RewardID:   reward.ID,
			CampaignID: event.CampaignID,
			UserID:     event.UserID,
			TenantID:   event.TenantID,
			AppID:      event.AppID,
			Config:     validatorConfig,
			Reward:     reward,
			Campaign:   campaign,
		}
		if err := s.validator.Validate(ctx, validationReq); err != nil {
			s.log.Warnf("validation failed: %v", err)
			return err
		}
	}

	// 4. 库存预占
	var reservationID string
	if s.iruc != nil {
		// 检查是否需要库存预占
		if validatorConfig != nil {
			if _, hasInventory := validatorConfig["INVENTORY"]; hasInventory {
				reservation := &InventoryReservation{
					ResourceID: reward.ID,
					CampaignID: event.CampaignID,
					UserID:     event.UserID,
					Quantity:   1,
					Status:     constants.InventoryReservationStatusPending,
					ExpireAt:   time.Now().Add(30 * time.Minute), // 30分钟过期
				}
				result, err := s.iruc.Reserve(ctx, reservation)
				if err != nil {
					return err
				}
				reservationID = result.ReservationID
			}
		}
	}

	// 5. 生成奖励内容
	generatorConfig, err := ParseGeneratorConfig(reward.GeneratorConfig)
	if err != nil {
		s.log.Warnf("failed to parse generator config: %v", err)
		generatorConfig = nil
	}

	generationReq := &GenerationRequest{
		RewardID:   reward.ID,
		RewardType: reward.RewardType,
		UserID:     event.UserID,
		Config:     generatorConfig,
		Reward:     reward,
	}

	content, err := s.generator.Generate(ctx, generationReq)
	if err != nil {
		s.log.Errorf("failed to generate reward content: %v", err)
		return err
	}

	// 6. 创建奖励发放记录
	grant := &RewardGrant{
		RewardID:        reward.ID,
		RewardName:      reward.Name,
		RewardType:      reward.RewardType,
		RewardVersion:   reward.Version,
		ContentSnapshot: content, // 使用生成的内容
		GeneratorConfig: reward.GeneratorConfig,
		CampaignID:      event.CampaignID,
		CampaignName:    event.CampaignName,
		TaskID:          task.ID,
		TaskName:        task.Name,
		TenantID:        event.TenantID,
		AppID:           event.AppID,
		UserID:          event.UserID,
		Status:          constants.RewardGrantStatusGenerated, // 已生成
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// 7. 计算过期时间
	if reward.ValidDays > 0 {
		expireTime := time.Now().AddDate(0, 0, reward.ValidDays)
		grant.ExpireTime = &expireTime
	}

	// 8. 保存发放记录
	if _, err := s.guc.Create(ctx, grant); err != nil {
		// 如果预占了库存，需要回滚
		if reservationID != "" && s.iruc != nil {
			_ = s.iruc.Cancel(ctx, reservationID)
		}
		return err
	}

	// 9. 更新完成日志的 GrantID
	log.GrantID = grant.GrantID
	_, _ = s.tcluc.Create(ctx, log)

	// 10. 执行实际发放
	distributorConfig, err := ParseDistributorConfig(reward.DistributorConfig)
	if err != nil {
		s.log.Warnf("failed to parse distributor config: %v", err)
		distributorConfig = nil
	}

	distributionReq := &DistributionRequest{
		GrantID:    grant.GrantID,
		RewardID:   reward.ID,
		RewardType: reward.RewardType,
		UserID:     event.UserID,
		Content:    content,
		Config:     distributorConfig,
	}

	if err := s.distributor.Distribute(ctx, distributionReq); err != nil {
		s.log.Errorf("failed to distribute reward: %v", err)
		grant.ErrorMessage = err.Error()
		_, _ = s.guc.Update(ctx, grant)
		return err
	}

	// 11. 更新状态为已发放
	grant.Status = constants.RewardGrantStatusDistributed
	now := time.Now()
	grant.DistributedAt = &now
	_, err = s.guc.Update(ctx, grant)

	// 12. 确认库存预占
	if reservationID != "" && s.iruc != nil {
		_ = s.iruc.Confirm(ctx, reservationID)
	}

	return err
}

// TaskEvent 任务事件
type TaskEvent struct {
	EventType    string                 // 事件类型，如：USER_REGISTER, ORDER_PAID
	UserID       int64                  // 用户ID
	TenantID     string                 // 租户ID
	AppID        string                 // 应用ID
	CampaignID   string                 // 活动ID（可选）
	CampaignName string                 // 活动名称（可选）
	EventData    map[string]interface{} // 事件数据
	Timestamp    time.Time              // 事件时间
}

// TaskEventMessage 任务事件消息（用于 RocketMQ）
type TaskEventMessage struct {
	EventType    string                 `json:"event_type"`
	UserID       int64                  `json:"user_id"`
	TenantID     string                 `json:"tenant_id"`
	AppID        string                 `json:"app_id"`
	CampaignID   string                 `json:"campaign_id,omitempty"`
	CampaignName string                 `json:"campaign_name,omitempty"`
	EventData    map[string]interface{} `json:"event_data"`
	Timestamp    string                 `json:"timestamp"`
}
