package service

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/biz"
	"marketing-service/internal/metrics"
)

// MarketingService 营销服务
type MarketingService struct {
	v1.UnimplementedMarketingServer

	uc    *biz.CampaignUseCase
	ruc   *biz.RewardUseCase
	guc   *biz.RewardGrantUseCase
	tuc   *biz.TaskUseCase
	auc   *biz.AudienceUseCase
	rdeuc *biz.RedeemCodeUseCase
	iruc  *biz.InventoryReservationUseCase
	tcluc *biz.TaskCompletionLogUseCase
	ctuc  *biz.CampaignTaskUseCase
	tts   *biz.TaskTriggerService
	log   *log.Helper
}

// NewMarketingService 创建营销服务
func NewMarketingService(
	uc *biz.CampaignUseCase,
	ruc *biz.RewardUseCase,
	guc *biz.RewardGrantUseCase,
	tuc *biz.TaskUseCase,
	auc *biz.AudienceUseCase,
	rdeuc *biz.RedeemCodeUseCase,
	iruc *biz.InventoryReservationUseCase,
	tcluc *biz.TaskCompletionLogUseCase,
	ctuc *biz.CampaignTaskUseCase,
	tts *biz.TaskTriggerService,
	logger log.Logger,
) *MarketingService {
	return &MarketingService{
		uc:    uc,
		ruc:   ruc,
		guc:   guc,
		tuc:   tuc,
		auc:   auc,
		rdeuc: rdeuc,
		iruc:  iruc,
		tcluc: tcluc,
		ctuc:  ctuc,
		tts:   tts,
		log:   log.NewHelper(logger),
	}
}

// CreateCampaign 创建营销活动
func (s *MarketingService) CreateCampaign(ctx context.Context, req *v1.CreateCampaignRequest) (*v1.CreateCampaignReply, error) {
	// 解析时间
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, err
	}

	campaign := &biz.Campaign{
		TenantID:        req.TenantId,
		AppID:           req.ProductCode, // 使用 product_code 作为 app_id
		Name:            req.CampaignName,
		Type:            req.CampaignType,
		StartTime:       startTime,
		EndTime:         endTime,
		AudienceConfig:  "", // TODO: 从 rule_config 中提取
		ValidatorConfig: "", // TODO: 从 rule_config 中提取
		Status:          "ACTIVE",
		Description:     "",
		CreatedBy:       req.CreatedBy,
	}

	result, err := s.uc.Create(ctx, campaign)
	if err != nil {
		s.log.Errorf("failed to create campaign: %v", err)
		return nil, err
	}

	return &v1.CreateCampaignReply{
		Campaign: s.toProtoCampaign(result),
	}, nil
}

// GetCampaign 获取营销活动
func (s *MarketingService) GetCampaign(ctx context.Context, req *v1.GetCampaignRequest) (*v1.GetCampaignReply, error) {
	campaign, err := s.uc.Get(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to get campaign: %v", err)
		return nil, err
	}
	if campaign == nil {
		return &v1.GetCampaignReply{}, nil
	}

	return &v1.GetCampaignReply{
		Campaign: s.toProtoCampaign(campaign),
	}, nil
}

// ListCampaigns 列出营销活动
func (s *MarketingService) ListCampaigns(ctx context.Context, req *v1.ListCampaignsRequest) (*v1.ListCampaignsReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	campaigns, total, err := s.uc.List(ctx, req.TenantId, req.ProductCode, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list campaigns: %v", err)
		return nil, err
	}

	protoCampaigns := make([]*v1.Campaign, 0, len(campaigns))
	for _, c := range campaigns {
		protoCampaigns = append(protoCampaigns, s.toProtoCampaign(c))
	}

	return &v1.ListCampaignsReply{
		Campaigns: protoCampaigns,
		Total:     int32(total),
	}, nil
}

// UpdateCampaign 更新营销活动
func (s *MarketingService) UpdateCampaign(ctx context.Context, req *v1.UpdateCampaignRequest) (*v1.UpdateCampaignReply, error) {
	campaign, err := s.uc.Get(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to get campaign: %v", err)
		return nil, err
	}
	if campaign == nil {
		return nil, err
	}

	// 更新字段
	if req.CampaignName != "" {
		campaign.Name = req.CampaignName
	}
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			campaign.StartTime = startTime
		}
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			campaign.EndTime = endTime
		}
	}
	if req.Status >= 0 {
		// TODO: 转换状态
	}

	result, err := s.uc.Update(ctx, campaign)
	if err != nil {
		s.log.Errorf("failed to update campaign: %v", err)
		return nil, err
	}

	return &v1.UpdateCampaignReply{
		Campaign: s.toProtoCampaign(result),
	}, nil
}

// DeleteCampaign 删除营销活动
func (s *MarketingService) DeleteCampaign(ctx context.Context, req *v1.DeleteCampaignRequest) (*v1.DeleteCampaignReply, error) {
	err := s.uc.Delete(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to delete campaign: %v", err)
		return nil, err
	}

	return &v1.DeleteCampaignReply{
		Success: true,
	}, nil
}

// toProtoCampaign 转换为 Proto Campaign
func (s *MarketingService) toProtoCampaign(c *biz.Campaign) *v1.Campaign {
	status := int32(0)
	if c.Status == "ACTIVE" {
		status = 1
	} else if c.Status == "ENDED" {
		status = 2
	}

	return &v1.Campaign{
		CampaignId:   c.ID,
		CampaignName: c.Name,
		TenantId:     c.TenantID,
		ProductCode:  c.AppID,
		CampaignType: c.Type,
		StartTime:    c.StartTime.Format(time.RFC3339),
		EndTime:      c.EndTime.Format(time.RFC3339),
		Status:       status,
		CreatedBy:    c.CreatedBy,
		CreatedAt:    timestamppb.New(c.CreatedAt).String(),
		UpdatedAt:    timestamppb.New(c.UpdatedAt).String(),
	}
}

// GenerateRedeemCodes 生成兑换码
func (s *MarketingService) GenerateRedeemCodes(ctx context.Context, req *v1.GenerateRedeemCodesRequest) (*v1.GenerateRedeemCodesReply, error) {
	// 记录指标
	metrics.GetMetrics().RedeemCodeGeneratedTotal.Add(float64(req.Count))
	// 获取活动信息
	campaign, err := s.uc.Get(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to get campaign: %v", err)
		return nil, err
	}
	if campaign == nil {
		return nil, err
	}

	// 生成批次ID
	batchID := time.Now().Format("20060102150405") + "-" + req.CampaignId[:8]

	// 解析过期时间
	var expireAt *time.Time
	if req.ExpireTime != "" {
		exp, err := time.Parse(time.RFC3339, req.ExpireTime)
		if err == nil {
			expireAt = &exp
		}
	}

	// 生成兑换码（性能优化：预分配容量，批量生成）
	codes := make([]*biz.RedeemCode, 0, req.Count)
	now := time.Now()

	// 使用 goroutine 池或批量生成以提高性能（对于大量数据）
	// 这里简化处理，直接循环生成
	for i := int32(0); i < req.Count; i++ {
		code := generateCode(req.CodeType)
		rc := &biz.RedeemCode{
			Code:         code,
			TenantID:     campaign.TenantID,
			AppID:        campaign.AppID,
			CampaignID:   req.CampaignId,
			CampaignName: campaign.Name,
			BatchID:      batchID,
			Status:       "ACTIVE",
			ExpireAt:     expireAt,
			CreatedAt:    now, // 使用相同的时间戳，减少系统调用
			UpdatedAt:    now,
		}
		codes = append(codes, rc)
	}

	// 批量创建
	if err := s.rdeuc.BatchCreate(ctx, codes); err != nil {
		s.log.Errorf("failed to batch create redeem codes: %v", err)
		return nil, err
	}

	// 返回样例码（最多10个）
	sampleCodes := make([]string, 0, 10)
	for i, code := range codes {
		if i >= 10 {
			break
		}
		sampleCodes = append(sampleCodes, code.Code)
	}

	return &v1.GenerateRedeemCodesReply{
		BatchId:     batchID,
		TotalCount:  req.Count,
		SampleCodes: sampleCodes,
	}, nil
}

// RedeemCode 兑换码核销
func (s *MarketingService) RedeemCode(ctx context.Context, req *v1.RedeemCodeRequest) (*v1.RedeemCodeReply, error) {
	// 记录指标（在成功时）
	defer func() {
		// 如果成功，会在后面记录
	}()
	// 查找兑换码
	code, err := s.rdeuc.GetByCode(ctx, req.Code, req.ProductCode)
	if err != nil {
		s.log.Errorf("failed to get redeem code: %v", err)
		return nil, err
	}
	if code == nil {
		return &v1.RedeemCodeReply{
			Success: false,
			Message: "redeem code not found",
		}, nil
	}

	// 检查状态
	if code.Status != "ACTIVE" {
		return &v1.RedeemCodeReply{
			Success:  false,
			Message:  "redeem code is not active",
			CodeInfo: s.toProtoRedeemCode(code),
		}, nil
	}

	// 检查是否过期
	if code.ExpireAt != nil && code.ExpireAt.Before(time.Now()) {
		// 更新状态为过期
		_ = s.rdeuc.UpdateStatus(ctx, req.Code, code.TenantID, "EXPIRED")
		return &v1.RedeemCodeReply{
			Success:  false,
			Message:  "redeem code expired",
			CodeInfo: s.toProtoRedeemCode(code),
		}, nil
	}

	// 执行核销
	if err := s.rdeuc.Redeem(ctx, req.Code, code.TenantID, req.UserId); err != nil {
		s.log.Errorf("failed to redeem code: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().RedeemCodeRedeemedTotal.Inc()

	// 重新获取兑换码信息
	code, _ = s.rdeuc.GetByCode(ctx, req.Code, code.TenantID)

	return &v1.RedeemCodeReply{
		Success: true,
		Message: "redeem success",
		RewardDetail: map[string]string{
			"campaign_id":   code.CampaignID,
			"campaign_name": code.CampaignName,
			"reward_id":     code.RewardID,
			"reward_name":   code.RewardName,
		},
		CodeInfo: s.toProtoRedeemCode(code),
	}, nil
}

// AssignRedeemCode 分配兑换码
func (s *MarketingService) AssignRedeemCode(ctx context.Context, req *v1.AssignRedeemCodeRequest) (*v1.AssignRedeemCodeReply, error) {
	// 查找兑换码（需要从请求中获取 tenant_id，这里简化处理）
	// 实际应该从 context 或请求中获取 tenant_id
	code, err := s.rdeuc.GetByCode(ctx, req.Code, "")
	if err != nil {
		s.log.Errorf("failed to get redeem code: %v", err)
		return nil, err
	}
	if code == nil {
		return &v1.AssignRedeemCodeReply{
			Success: false,
			Message: "redeem code not found",
		}, nil
	}

	// 检查状态
	if code.Status != "ACTIVE" {
		return &v1.AssignRedeemCodeReply{
			Success:  false,
			Message:  "redeem code is not active",
			CodeInfo: s.toProtoRedeemCode(code),
		}, nil
	}

	// 更新拥有者
	code.OwnerUserID = &req.UserId
	code.UpdatedAt = time.Now()
	if _, err := s.rdeuc.Create(ctx, code); err != nil {
		s.log.Errorf("failed to assign redeem code: %v", err)
		return nil, err
	}

	return &v1.AssignRedeemCodeReply{
		Success:  true,
		Message:  "assign success",
		CodeInfo: s.toProtoRedeemCode(code),
	}, nil
}

// ListRedeemCodes 列出兑换码
func (s *MarketingService) ListRedeemCodes(ctx context.Context, req *v1.ListRedeemCodesRequest) (*v1.ListRedeemCodesReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	status := ""
	if req.Status >= 0 {
		statusMap := map[int32]string{
			0: "ACTIVE",
			1: "REDEEMED",
			2: "EXPIRED",
			3: "REVOKED",
		}
		status = statusMap[req.Status]
	}

	codes, total, err := s.rdeuc.List(ctx, req.TenantId, req.ProductCode, req.CampaignId, "", req.UserId, status, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list redeem codes: %v", err)
		return nil, err
	}

	protoCodes := make([]*v1.RedeemCode, 0, len(codes))
	for _, c := range codes {
		protoCodes = append(protoCodes, s.toProtoRedeemCode(c))
	}

	return &v1.ListRedeemCodesReply{
		Codes: protoCodes,
		Total: int32(total),
	}, nil
}

// GetRedeemCode 获取兑换码
func (s *MarketingService) GetRedeemCode(ctx context.Context, req *v1.GetRedeemCodeRequest) (*v1.GetRedeemCodeReply, error) {
	// 需要从 context 或请求中获取 tenant_id，这里简化处理
	code, err := s.rdeuc.GetByCode(ctx, req.Code, "")
	if err != nil {
		s.log.Errorf("failed to get redeem code: %v", err)
		return nil, err
	}
	if code == nil {
		return &v1.GetRedeemCodeReply{}, nil
	}

	return &v1.GetRedeemCodeReply{
		Code: s.toProtoRedeemCode(code),
	}, nil
}

// toProtoRedeemCode 转换为 Proto RedeemCode
func (s *MarketingService) toProtoRedeemCode(rc *biz.RedeemCode) *v1.RedeemCode {
	status := int32(0)
	switch rc.Status {
	case "ACTIVE":
		status = 0
	case "REDEEMED":
		status = 2
	case "EXPIRED":
		status = 3
	case "REVOKED":
		status = 3
	}

	var validUntil, redemptionAt *timestamppb.Timestamp
	if rc.ExpireAt != nil {
		validUntil = timestamppb.New(*rc.ExpireAt)
	}
	if rc.RedeemedAt != nil {
		redemptionAt = timestamppb.New(*rc.RedeemedAt)
	}

	return &v1.RedeemCode{
		Code:         rc.Code,
		CampaignId:   rc.CampaignID,
		TenantId:     rc.TenantID,
		ProductCode:  rc.AppID,
		UserId:       getInt64Value(rc.OwnerUserID, rc.RedeemedBy),
		Status:       status,
		ValidUntil:   validUntil,
		RedemptionAt: redemptionAt,
		CreatedAt:    timestamppb.New(rc.CreatedAt),
		UpdatedAt:    timestamppb.New(rc.UpdatedAt),
	}
}

// generateCode 生成兑换码
func generateCode(codeType string) string {
	// 简单的兑换码生成逻辑，实际应该根据 codeType 生成不同类型的码
	chars := "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 8)
	for i := range code {
		code[i] = chars[time.Now().UnixNano()%int64(len(chars))]
	}
	return string(code)
}

// getInt64Value 获取 int64 值（优先返回第一个非 nil 的值）
func getInt64Value(values ...*int64) int64 {
	for _, v := range values {
		if v != nil {
			return *v
		}
	}
	return 0
}

// ========== Reward API ==========

// CreateReward 创建奖励
func (s *MarketingService) CreateReward(ctx context.Context, req *v1.CreateRewardRequest) (*v1.CreateRewardReply, error) {
	reward := &biz.Reward{
		TenantID:          req.TenantId,
		AppID:             req.ProductCode,
		RewardType:        req.RewardType,
		Name:              req.Name,
		ContentConfig:     req.ContentConfig,
		GeneratorConfig:   req.GeneratorConfig,
		DistributorConfig: req.DistributorConfig,
		ValidatorConfig:   req.ValidatorConfig,
		ValidDays:         int(req.ValidDays),
		Description:       req.Description,
		CreatedBy:         req.CreatedBy,
	}

	result, err := s.ruc.Create(ctx, reward)
	if err != nil {
		s.log.Errorf("failed to create reward: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().RewardCreatedTotal.Inc()

	return &v1.CreateRewardReply{
		Reward: s.toProtoReward(result),
	}, nil
}

// GetReward 获取奖励
func (s *MarketingService) GetReward(ctx context.Context, req *v1.GetRewardRequest) (*v1.GetRewardReply, error) {
	reward, err := s.ruc.Get(ctx, req.RewardId)
	if err != nil {
		s.log.Errorf("failed to get reward: %v", err)
		return nil, err
	}
	if reward == nil {
		return &v1.GetRewardReply{}, nil
	}

	return &v1.GetRewardReply{
		Reward: s.toProtoReward(reward),
	}, nil
}

// ListRewards 列出奖励
func (s *MarketingService) ListRewards(ctx context.Context, req *v1.ListRewardsRequest) (*v1.ListRewardsReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	rewards, total, err := s.ruc.List(ctx, req.TenantId, req.ProductCode, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list rewards: %v", err)
		return nil, err
	}

	protoRewards := make([]*v1.Reward, 0, len(rewards))
	for _, r := range rewards {
		protoRewards = append(protoRewards, s.toProtoReward(r))
	}

	return &v1.ListRewardsReply{
		Rewards: protoRewards,
		Total:   int32(total),
	}, nil
}

// UpdateReward 更新奖励
func (s *MarketingService) UpdateReward(ctx context.Context, req *v1.UpdateRewardRequest) (*v1.UpdateRewardReply, error) {
	reward, err := s.ruc.Get(ctx, req.RewardId)
	if err != nil {
		s.log.Errorf("failed to get reward: %v", err)
		return nil, err
	}
	if reward == nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		reward.Name = req.Name
	}
	if req.ContentConfig != "" {
		reward.ContentConfig = req.ContentConfig
	}
	if req.GeneratorConfig != "" {
		reward.GeneratorConfig = req.GeneratorConfig
	}
	if req.DistributorConfig != "" {
		reward.DistributorConfig = req.DistributorConfig
	}
	if req.ValidatorConfig != "" {
		reward.ValidatorConfig = req.ValidatorConfig
	}
	if req.ValidDays > 0 {
		reward.ValidDays = int(req.ValidDays)
	}
	if req.Status != "" {
		reward.Status = req.Status
	}
	if req.Description != "" {
		reward.Description = req.Description
	}

	result, err := s.ruc.Update(ctx, reward)
	if err != nil {
		s.log.Errorf("failed to update reward: %v", err)
		return nil, err
	}

	return &v1.UpdateRewardReply{
		Reward: s.toProtoReward(result),
	}, nil
}

// DeleteReward 删除奖励
func (s *MarketingService) DeleteReward(ctx context.Context, req *v1.DeleteRewardRequest) (*v1.DeleteRewardReply, error) {
	err := s.ruc.Delete(ctx, req.RewardId)
	if err != nil {
		s.log.Errorf("failed to delete reward: %v", err)
		return nil, err
	}

	return &v1.DeleteRewardReply{
		Success: true,
	}, nil
}

// toProtoReward 转换为 Proto Reward
func (s *MarketingService) toProtoReward(r *biz.Reward) *v1.Reward {
	return &v1.Reward{
		RewardId:          r.ID,
		TenantId:          r.TenantID,
		ProductCode:       r.AppID,
		RewardType:        r.RewardType,
		Name:              r.Name,
		ContentConfig:     r.ContentConfig,
		GeneratorConfig:   r.GeneratorConfig,
		DistributorConfig: r.DistributorConfig,
		ValidatorConfig:   r.ValidatorConfig,
		Version:           int32(r.Version),
		ValidDays:         int32(r.ValidDays),
		Status:            r.Status,
		Description:       r.Description,
		CreatedBy:         r.CreatedBy,
		CreatedAt:         timestamppb.New(r.CreatedAt),
		UpdatedAt:         timestamppb.New(r.UpdatedAt),
	}
}

// ========== Task API ==========

// CreateTask 创建任务
func (s *MarketingService) CreateTask(ctx context.Context, req *v1.CreateTaskRequest) (*v1.CreateTaskReply, error) {
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		return nil, err
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		return nil, err
	}

	task := &biz.Task{
		TenantID:        req.TenantId,
		AppID:           req.ProductCode,
		Name:            req.Name,
		TaskType:        req.TaskType,
		TriggerConfig:   req.TriggerConfig,
		ConditionConfig: req.ConditionConfig,
		RewardID:        req.RewardId,
		StartTime:       startTime,
		EndTime:         endTime,
		MaxCount:        int(req.MaxCount),
		Description:     req.Description,
		CreatedBy:       req.CreatedBy,
	}

	result, err := s.tuc.Create(ctx, task)
	if err != nil {
		s.log.Errorf("failed to create task: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().TaskCreatedTotal.Inc()

	return &v1.CreateTaskReply{
		Task: s.toProtoTask(result),
	}, nil
}

// GetTask 获取任务
func (s *MarketingService) GetTask(ctx context.Context, req *v1.GetTaskRequest) (*v1.GetTaskReply, error) {
	task, err := s.tuc.Get(ctx, req.TaskId)
	if err != nil {
		s.log.Errorf("failed to get task: %v", err)
		return nil, err
	}
	if task == nil {
		return &v1.GetTaskReply{}, nil
	}

	return &v1.GetTaskReply{
		Task: s.toProtoTask(task),
	}, nil
}

// ListTasks 列出任务
func (s *MarketingService) ListTasks(ctx context.Context, req *v1.ListTasksRequest) (*v1.ListTasksReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	tasks, total, err := s.tuc.List(ctx, req.TenantId, req.ProductCode, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list tasks: %v", err)
		return nil, err
	}

	protoTasks := make([]*v1.Task, 0, len(tasks))
	for _, t := range tasks {
		protoTasks = append(protoTasks, s.toProtoTask(t))
	}

	return &v1.ListTasksReply{
		Tasks: protoTasks,
		Total: int32(total),
	}, nil
}

// UpdateTask 更新任务
func (s *MarketingService) UpdateTask(ctx context.Context, req *v1.UpdateTaskRequest) (*v1.UpdateTaskReply, error) {
	task, err := s.tuc.Get(ctx, req.TaskId)
	if err != nil {
		s.log.Errorf("failed to get task: %v", err)
		return nil, err
	}
	if task == nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		task.Name = req.Name
	}
	if req.TriggerConfig != "" {
		task.TriggerConfig = req.TriggerConfig
	}
	if req.ConditionConfig != "" {
		task.ConditionConfig = req.ConditionConfig
	}
	if req.RewardId != "" {
		task.RewardID = req.RewardId
	}
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			task.StartTime = startTime
		}
	}
	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			task.EndTime = endTime
		}
	}
	if req.MaxCount > 0 {
		task.MaxCount = int(req.MaxCount)
	}
	if req.Status != "" {
		task.Status = req.Status
	}
	if req.Description != "" {
		task.Description = req.Description
	}

	result, err := s.tuc.Update(ctx, task)
	if err != nil {
		s.log.Errorf("failed to update task: %v", err)
		return nil, err
	}

	return &v1.UpdateTaskReply{
		Task: s.toProtoTask(result),
	}, nil
}

// DeleteTask 删除任务
func (s *MarketingService) DeleteTask(ctx context.Context, req *v1.DeleteTaskRequest) (*v1.DeleteTaskReply, error) {
	err := s.tuc.Delete(ctx, req.TaskId)
	if err != nil {
		s.log.Errorf("failed to delete task: %v", err)
		return nil, err
	}

	return &v1.DeleteTaskReply{
		Success: true,
	}, nil
}

// ListTasksByCampaign 根据活动列出任务
func (s *MarketingService) ListTasksByCampaign(ctx context.Context, req *v1.ListTasksByCampaignRequest) (*v1.ListTasksByCampaignReply, error) {
	tasks, err := s.tuc.ListByCampaign(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to list tasks by campaign: %v", err)
		return nil, err
	}

	protoTasks := make([]*v1.Task, 0, len(tasks))
	for _, t := range tasks {
		protoTasks = append(protoTasks, s.toProtoTask(t))
	}

	return &v1.ListTasksByCampaignReply{
		Tasks: protoTasks,
	}, nil
}

// toProtoTask 转换为 Proto Task
func (s *MarketingService) toProtoTask(t *biz.Task) *v1.Task {
	return &v1.Task{
		TaskId:          t.ID,
		TenantId:        t.TenantID,
		ProductCode:     t.AppID,
		Name:            t.Name,
		TaskType:        t.TaskType,
		TriggerConfig:   t.TriggerConfig,
		ConditionConfig: t.ConditionConfig,
		RewardId:        t.RewardID,
		Status:          t.Status,
		StartTime:       t.StartTime.Format(time.RFC3339),
		EndTime:         t.EndTime.Format(time.RFC3339),
		MaxCount:        int32(t.MaxCount),
		Description:     t.Description,
		CreatedBy:       t.CreatedBy,
		CreatedAt:       timestamppb.New(t.CreatedAt),
		UpdatedAt:       timestamppb.New(t.UpdatedAt),
	}
}

// ========== Audience API ==========

// CreateAudience 创建受众
func (s *MarketingService) CreateAudience(ctx context.Context, req *v1.CreateAudienceRequest) (*v1.CreateAudienceReply, error) {
	audience := &biz.Audience{
		TenantID:     req.TenantId,
		AppID:        req.ProductCode,
		Name:         req.Name,
		AudienceType: req.AudienceType,
		RuleConfig:   req.RuleConfig,
		Description:  req.Description,
		CreatedBy:    req.CreatedBy,
	}

	result, err := s.auc.Create(ctx, audience)
	if err != nil {
		s.log.Errorf("failed to create audience: %v", err)
		return nil, err
	}

	return &v1.CreateAudienceReply{
		Audience: s.toProtoAudience(result),
	}, nil
}

// GetAudience 获取受众
func (s *MarketingService) GetAudience(ctx context.Context, req *v1.GetAudienceRequest) (*v1.GetAudienceReply, error) {
	audience, err := s.auc.Get(ctx, req.AudienceId)
	if err != nil {
		s.log.Errorf("failed to get audience: %v", err)
		return nil, err
	}
	if audience == nil {
		return &v1.GetAudienceReply{}, nil
	}

	return &v1.GetAudienceReply{
		Audience: s.toProtoAudience(audience),
	}, nil
}

// ListAudiences 列出受众
func (s *MarketingService) ListAudiences(ctx context.Context, req *v1.ListAudiencesRequest) (*v1.ListAudiencesReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	audiences, total, err := s.auc.List(ctx, req.TenantId, req.ProductCode, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list audiences: %v", err)
		return nil, err
	}

	protoAudiences := make([]*v1.Audience, 0, len(audiences))
	for _, a := range audiences {
		protoAudiences = append(protoAudiences, s.toProtoAudience(a))
	}

	return &v1.ListAudiencesReply{
		Audiences: protoAudiences,
		Total:     int32(total),
	}, nil
}

// UpdateAudience 更新受众
func (s *MarketingService) UpdateAudience(ctx context.Context, req *v1.UpdateAudienceRequest) (*v1.UpdateAudienceReply, error) {
	audience, err := s.auc.Get(ctx, req.AudienceId)
	if err != nil {
		s.log.Errorf("failed to get audience: %v", err)
		return nil, err
	}
	if audience == nil {
		return nil, err
	}

	// 更新字段
	if req.Name != "" {
		audience.Name = req.Name
	}
	if req.RuleConfig != "" {
		audience.RuleConfig = req.RuleConfig
	}
	if req.Status != "" {
		audience.Status = req.Status
	}
	if req.Description != "" {
		audience.Description = req.Description
	}

	result, err := s.auc.Update(ctx, audience)
	if err != nil {
		s.log.Errorf("failed to update audience: %v", err)
		return nil, err
	}

	return &v1.UpdateAudienceReply{
		Audience: s.toProtoAudience(result),
	}, nil
}

// DeleteAudience 删除受众
func (s *MarketingService) DeleteAudience(ctx context.Context, req *v1.DeleteAudienceRequest) (*v1.DeleteAudienceReply, error) {
	err := s.auc.Delete(ctx, req.AudienceId)
	if err != nil {
		s.log.Errorf("failed to delete audience: %v", err)
		return nil, err
	}

	return &v1.DeleteAudienceReply{
		Success: true,
	}, nil
}

// toProtoAudience 转换为 Proto Audience
func (s *MarketingService) toProtoAudience(a *biz.Audience) *v1.Audience {
	return &v1.Audience{
		AudienceId:   a.ID,
		TenantId:     a.TenantID,
		ProductCode:  a.AppID,
		Name:         a.Name,
		AudienceType: a.AudienceType,
		RuleConfig:   a.RuleConfig,
		Status:       a.Status,
		Description:  a.Description,
		CreatedBy:    a.CreatedBy,
		CreatedAt:    timestamppb.New(a.CreatedAt),
		UpdatedAt:    timestamppb.New(a.UpdatedAt),
	}
}

// ========== RewardGrant API ==========

// ListRewardGrants 列出奖励发放记录
func (s *MarketingService) ListRewardGrants(ctx context.Context, req *v1.ListRewardGrantsRequest) (*v1.ListRewardGrantsReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	grants, total, err := s.guc.List(ctx, req.TenantId, req.ProductCode, req.UserId, req.Status, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list reward grants: %v", err)
		return nil, err
	}

	protoGrants := make([]*v1.RewardGrant, 0, len(grants))
	for _, g := range grants {
		protoGrants = append(protoGrants, s.toProtoRewardGrant(g))
	}

	return &v1.ListRewardGrantsReply{
		Grants: protoGrants,
		Total:  int32(total),
	}, nil
}

// GetRewardGrant 获取奖励发放记录
func (s *MarketingService) GetRewardGrant(ctx context.Context, req *v1.GetRewardGrantRequest) (*v1.GetRewardGrantReply, error) {
	grant, err := s.guc.Get(ctx, req.GrantId)
	if err != nil {
		s.log.Errorf("failed to get reward grant: %v", err)
		return nil, err
	}
	if grant == nil {
		return &v1.GetRewardGrantReply{}, nil
	}

	return &v1.GetRewardGrantReply{
		Grant: s.toProtoRewardGrant(grant),
	}, nil
}

// UpdateRewardGrantStatus 更新发放状态
func (s *MarketingService) UpdateRewardGrantStatus(ctx context.Context, req *v1.UpdateRewardGrantStatusRequest) (*v1.UpdateRewardGrantStatusReply, error) {
	err := s.guc.UpdateStatus(ctx, req.GrantId, req.Status)
	if err != nil {
		s.log.Errorf("failed to update reward grant status: %v", err)
		return nil, err
	}

	return &v1.UpdateRewardGrantStatusReply{
		Success: true,
	}, nil
}

// toProtoRewardGrant 转换为 Proto RewardGrant
func (s *MarketingService) toProtoRewardGrant(g *biz.RewardGrant) *v1.RewardGrant {
	var reservedAt, distributedAt, usedAt, expireTime *timestamppb.Timestamp
	if g.ReservedAt != nil {
		reservedAt = timestamppb.New(*g.ReservedAt)
	}
	if g.DistributedAt != nil {
		distributedAt = timestamppb.New(*g.DistributedAt)
	}
	if g.UsedAt != nil {
		usedAt = timestamppb.New(*g.UsedAt)
	}
	if g.ExpireTime != nil {
		expireTime = timestamppb.New(*g.ExpireTime)
	}

	return &v1.RewardGrant{
		GrantId:         g.GrantID,
		RewardId:        g.RewardID,
		RewardName:      g.RewardName,
		RewardType:      g.RewardType,
		RewardVersion:   int32(g.RewardVersion),
		ContentSnapshot: g.ContentSnapshot,
		CampaignId:      g.CampaignID,
		CampaignName:    g.CampaignName,
		TaskId:          g.TaskID,
		TaskName:        g.TaskName,
		TenantId:        g.TenantID,
		ProductCode:     g.AppID,
		UserId:          g.UserID,
		Status:          g.Status,
		ReservedAt:      reservedAt,
		DistributedAt:   distributedAt,
		UsedAt:          usedAt,
		ExpireTime:      expireTime,
		ErrorMessage:    g.ErrorMessage,
		CreatedAt:       timestamppb.New(g.CreatedAt),
		UpdatedAt:       timestamppb.New(g.UpdatedAt),
	}
}

// ========== Task Event API ==========

// TriggerTaskEvent 触发任务事件
func (s *MarketingService) TriggerTaskEvent(ctx context.Context, req *v1.TriggerTaskEventRequest) (*v1.TriggerTaskEventReply, error) {
	// 转换事件数据
	eventData := make(map[string]interface{})
	for k, v := range req.EventData {
		eventData[k] = v
	}

	event := &biz.TaskEvent{
		EventType:    req.EventType,
		UserID:       req.UserId,
		TenantID:     req.TenantId,
		AppID:        req.ProductCode,
		CampaignID:   req.CampaignId,
		CampaignName: req.CampaignName,
		EventData:    eventData,
		Timestamp:    time.Now(),
	}

	err := s.tts.TriggerEvent(ctx, event)
	if err != nil {
		s.log.Errorf("failed to trigger task event: %v", err)
		return &v1.TriggerTaskEventReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.TriggerTaskEventReply{
		Success:        true,
		Message:        "task event triggered successfully",
		TasksTriggered: 0, // TODO: 从 TaskTriggerService 返回触发的任务数量
		RewardsIssued:  0, // TODO: 从 TaskTriggerService 返回发放的奖励数量
	}, nil
}

// ========== Inventory Management API ==========

// ReserveInventory 预占库存
func (s *MarketingService) ReserveInventory(ctx context.Context, req *v1.ReserveInventoryRequest) (*v1.ReserveInventoryReply, error) {
	expireMinutes := int(req.ExpireMinutes)
	if expireMinutes <= 0 {
		expireMinutes = 30 // 默认30分钟
	}

	reservation := &biz.InventoryReservation{
		ResourceID: req.ResourceId,
		CampaignID: req.CampaignId,
		UserID:     req.UserId,
		Quantity:   int(req.Quantity),
		ExpireAt:   time.Now().Add(time.Duration(expireMinutes) * time.Minute),
	}

	result, err := s.iruc.Reserve(ctx, reservation)
	if err != nil {
		s.log.Errorf("failed to reserve inventory: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().InventoryReservedTotal.Add(float64(req.Quantity))

	return &v1.ReserveInventoryReply{
		Reservation: s.toProtoInventoryReservation(result),
	}, nil
}

// ConfirmInventory 确认库存
func (s *MarketingService) ConfirmInventory(ctx context.Context, req *v1.ConfirmInventoryRequest) (*v1.ConfirmInventoryReply, error) {
	err := s.iruc.Confirm(ctx, req.ReservationId)
	if err != nil {
		s.log.Errorf("failed to confirm inventory: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().InventoryConfirmedTotal.Inc()

	return &v1.ConfirmInventoryReply{
		Success: true,
	}, nil
}

// CancelInventory 取消库存
func (s *MarketingService) CancelInventory(ctx context.Context, req *v1.CancelInventoryRequest) (*v1.CancelInventoryReply, error) {
	err := s.iruc.Cancel(ctx, req.ReservationId)
	if err != nil {
		s.log.Errorf("failed to cancel inventory: %v", err)
		return nil, err
	}

	// 记录指标
	metrics.GetMetrics().InventoryCancelledTotal.Inc()

	return &v1.CancelInventoryReply{
		Success: true,
	}, nil
}

// ListInventoryReservations 列出库存预占记录
func (s *MarketingService) ListInventoryReservations(ctx context.Context, req *v1.ListInventoryReservationsRequest) (*v1.ListInventoryReservationsReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	reservations, total, err := s.iruc.List(ctx, req.ResourceId, req.CampaignId, req.UserId, req.Status, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list inventory reservations: %v", err)
		return nil, err
	}

	protoReservations := make([]*v1.InventoryReservation, 0, len(reservations))
	for _, r := range reservations {
		protoReservations = append(protoReservations, s.toProtoInventoryReservation(r))
	}

	return &v1.ListInventoryReservationsReply{
		Reservations: protoReservations,
		Total:        int32(total),
	}, nil
}

// toProtoInventoryReservation 转换为 Proto InventoryReservation
func (s *MarketingService) toProtoInventoryReservation(ir *biz.InventoryReservation) *v1.InventoryReservation {
	return &v1.InventoryReservation{
		ReservationId: ir.ReservationID,
		ResourceId:    ir.ResourceID,
		CampaignId:    ir.CampaignID,
		UserId:        ir.UserID,
		Quantity:      int32(ir.Quantity),
		Status:        ir.Status,
		ExpireAt:      timestamppb.New(ir.ExpireAt),
		CreatedAt:     timestamppb.New(ir.CreatedAt),
		UpdatedAt:     timestamppb.New(ir.UpdatedAt),
	}
}

// ========== Task Completion Log API ==========

// ListTaskCompletionLogs 列出任务完成日志
func (s *MarketingService) ListTaskCompletionLogs(ctx context.Context, req *v1.ListTaskCompletionLogsRequest) (*v1.ListTaskCompletionLogsReply, error) {
	page := int(req.PageNum)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	logs, total, err := s.tcluc.List(ctx, req.TenantId, req.ProductCode, req.TaskId, req.CampaignId, req.UserId, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list task completion logs: %v", err)
		return nil, err
	}

	protoLogs := make([]*v1.TaskCompletionLog, 0, len(logs))
	for _, l := range logs {
		protoLogs = append(protoLogs, s.toProtoTaskCompletionLog(l))
	}

	return &v1.ListTaskCompletionLogsReply{
		Logs:  protoLogs,
		Total: int32(total),
	}, nil
}

// GetTaskCompletionStats 获取任务完成统计
func (s *MarketingService) GetTaskCompletionStats(ctx context.Context, req *v1.GetTaskCompletionStatsRequest) (*v1.GetTaskCompletionStatsReply, error) {
	// 获取总完成次数
	totalCompletions, err := s.tcluc.CountByTask(ctx, req.TaskId, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to count task completions: %v", err)
		return nil, err
	}

	// 获取唯一用户数
	uniqueUsers, err := s.tcluc.CountUniqueUsersByTask(ctx, req.TaskId, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to count unique users: %v", err)
		return nil, err
	}

	// 获取用户完成次数
	var userCompletions int64
	if req.UserId > 0 {
		count, err := s.tcluc.CountByTaskAndUser(ctx, req.TaskId, req.UserId)
		if err != nil {
			s.log.Errorf("failed to count task completions by user: %v", err)
			return nil, err
		}
		userCompletions = count
	}

	return &v1.GetTaskCompletionStatsReply{
		TaskId:           req.TaskId,
		TotalCompletions: int32(totalCompletions),
		UniqueUsers:      int32(uniqueUsers),
		UserCompletions:  userCompletions,
	}, nil
}

// toProtoTaskCompletionLog 转换为 Proto TaskCompletionLog
func (s *MarketingService) toProtoTaskCompletionLog(tcl *biz.TaskCompletionLog) *v1.TaskCompletionLog {
	return &v1.TaskCompletionLog{
		CompletionId: tcl.CompletionID,
		TaskId:       tcl.TaskID,
		TaskName:     tcl.TaskName,
		CampaignId:   tcl.CampaignID,
		CampaignName: tcl.CampaignName,
		UserId:       tcl.UserID,
		TenantId:     tcl.TenantID,
		ProductCode:  tcl.AppID,
		GrantId:      tcl.GrantID,
		ProgressData: tcl.ProgressData,
		TriggerEvent: tcl.TriggerEvent,
		CompletedAt:  timestamppb.New(tcl.CompletedAt),
		CreatedAt:    timestamppb.New(tcl.CreatedAt),
		UpdatedAt:    timestamppb.New(tcl.UpdatedAt),
	}
}

// ========== Campaign Task Management API ==========

// AddTaskToCampaign 将任务添加到活动
func (s *MarketingService) AddTaskToCampaign(ctx context.Context, req *v1.AddTaskToCampaignRequest) (*v1.AddTaskToCampaignReply, error) {
	result, err := s.ctuc.AddTaskToCampaign(ctx, req.CampaignId, req.TaskId, int(req.SortOrder), req.Config)
	if err != nil {
		s.log.Errorf("failed to add task to campaign: %v", err)
		return nil, err
	}

	return &v1.AddTaskToCampaignReply{
		CampaignTask: s.toProtoCampaignTask(result),
	}, nil
}

// RemoveTaskFromCampaign 从活动中移除任务
func (s *MarketingService) RemoveTaskFromCampaign(ctx context.Context, req *v1.RemoveTaskFromCampaignRequest) (*v1.RemoveTaskFromCampaignReply, error) {
	err := s.ctuc.RemoveTaskFromCampaign(ctx, req.CampaignId, req.TaskId)
	if err != nil {
		s.log.Errorf("failed to remove task from campaign: %v", err)
		return nil, err
	}

	return &v1.RemoveTaskFromCampaignReply{
		Success: true,
	}, nil
}

// ListCampaignTasks 列出活动的所有任务
func (s *MarketingService) ListCampaignTasks(ctx context.Context, req *v1.ListCampaignTasksRequest) (*v1.ListCampaignTasksReply, error) {
	campaignTasks, err := s.ctuc.ListCampaignTasks(ctx, req.CampaignId)
	if err != nil {
		s.log.Errorf("failed to list campaign tasks: %v", err)
		return nil, err
	}

	protoCampaignTasks := make([]*v1.CampaignTask, 0, len(campaignTasks))
	for _, ct := range campaignTasks {
		protoCampaignTasks = append(protoCampaignTasks, s.toProtoCampaignTask(ct))
	}

	return &v1.ListCampaignTasksReply{
		CampaignTasks: protoCampaignTasks,
	}, nil
}

// toProtoCampaignTask 转换为 Proto CampaignTask
func (s *MarketingService) toProtoCampaignTask(ct *biz.CampaignTask) *v1.CampaignTask {
	return &v1.CampaignTask{
		CampaignTaskId: ct.CampaignTaskID,
		CampaignId:     ct.CampaignID,
		TaskId:         ct.TaskID,
		Config:         ct.Config,
		SortOrder:      int32(ct.SortOrder),
		CreatedAt:      timestamppb.New(ct.CreatedAt),
		UpdatedAt:      timestamppb.New(ct.UpdatedAt),
	}
}
