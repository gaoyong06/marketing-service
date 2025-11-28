package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gaoyong06/middleground/marketing-service/internal/biz"
	v1 "github.com/gaoyong06/middleground/proto-repo/gen/go/platform/marketing_service/v1"
	"github.com/go-kratos/kratos/v2/log"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MarketingService is a marketing service.
type MarketingService struct {
	v1.UnimplementedMarketingServiceServer

	campaignUc   *biz.CampaignUsecase
	redeemCodeUc *biz.RedeemCodeUsecase
	log          *log.Helper
}

// NewMarketingService new a marketing service.
func NewMarketingService(campaignUc *biz.CampaignUsecase, redeemCodeUc *biz.RedeemCodeUsecase, logger log.Logger) *MarketingService {
	return &MarketingService{
		campaignUc:   campaignUc,
		redeemCodeUc: redeemCodeUc,
		log:          log.NewHelper(logger),
	}
}

// CreateCampaign 创建营销活动
func (s *MarketingService) CreateCampaign(ctx context.Context, req *v1.CreateCampaignRequest) (*v1.CreateCampaignReply, error) {
	if req.TenantId == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "tenant_id is required")
	}

	if req.CampaignName == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "campaign_name is required")
	}

	// 构建业务对象
	campaign := &biz.Campaign{
		TenantID:     req.TenantId,
		CampaignName: req.CampaignName,
		CampaignType: req.CampaignType,
		Status:       0, // 默认为未开始状态
	}

	// 设置时间
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

	// 设置规则配置
	if len(req.RuleConfig) > 0 {
		// 将 map[string]string 转换为 map[string]interface{}
		config := make(map[string]interface{})
		for k, v := range req.RuleConfig {
			config[k] = v
		}

		// 创建规则
		campaign.Rules = make([]*biz.CampaignRule, 0, 1)
		campaign.Rules = append(campaign.Rules, &biz.CampaignRule{
			RuleType: "default",
			Config:   config,
		})
	}

	// 创建活动
	result, err := s.campaignUc.CreateCampaign(ctx, campaign)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "create campaign failed: %v", err)
	}

	// 构建响应
	reply := &v1.CreateCampaignReply{
		Campaign: &v1.Campaign{
			CampaignId:   result.CampaignID,
			TenantId:     campaign.TenantID,
			CampaignName: campaign.CampaignName,
			CampaignType: campaign.CampaignType,
			Status:       campaign.Status,
		},
	}

	return reply, nil
}

// UpdateCampaign 更新营销活动
func (s *MarketingService) UpdateCampaign(ctx context.Context, req *v1.UpdateCampaignRequest) (*v1.UpdateCampaignReply, error) {
	if req.CampaignId == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "campaign_id is required")
	}

	// 先获取现有活动
	existing, err := s.campaignUc.GetCampaign(ctx, req.CampaignId)
	if err != nil {
		return nil, status.Errorf(grpccodes.NotFound, "campaign not found: %v", err)
	}

	// 更新字段
	if req.CampaignName != "" {
		existing.CampaignName = req.CampaignName
	}

	if req.Status != 0 {
		existing.Status = int32(req.Status)
	}

	// Description 字段在 proto 中未定义，所以不更新 Description

	// 设置时间
	// 在 proto 中，时间字段是 string 类型，需要解析
	if req.StartTime != "" {
		startTime, err := time.Parse(time.RFC3339, req.StartTime)
		if err == nil {
			existing.StartTime = startTime
		} else {
			s.log.Warnf("failed to parse start_time: %v", err)
		}
	}

	if req.EndTime != "" {
		endTime, err := time.Parse(time.RFC3339, req.EndTime)
		if err == nil {
			existing.EndTime = endTime
		} else {
			s.log.Warnf("failed to parse end_time: %v", err)
		}
	}

	// 设置规则
	if len(req.RuleConfig) > 0 {
		// 在 proto 中，规则配置是作为 map<string, string> 存储的
		// 这里我们将其转换为一个规则对象
		existing.Rules = make([]*biz.CampaignRule, 0, 1)
		// 需要将 map[string]string 转换为 map[string]interface{}
		configMap := make(map[string]interface{})
		for k, v := range req.RuleConfig {
			configMap[k] = v
		}

		existing.Rules = append(existing.Rules, &biz.CampaignRule{
			RuleType: "default", // 使用默认类型
			Config:   configMap,
		})
	}

	// 更新活动
	// 使用 _ 忽略返回的活动对象，因为我们只需要知道更新是否成功
	_, err = s.campaignUc.UpdateCampaign(ctx, existing)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "update campaign failed: %v", err)
	}

	// 构建响应
	// 根据 proto 定义，需要返回更新后的活动信息
	// 由于我们已经忽略了 UpdateCampaign 的返回值，所以需要再次获取活动信息
	updatedCampaign, err := s.campaignUc.GetCampaign(ctx, existing.CampaignID)
	if err != nil {
		s.log.Warnf("failed to get updated campaign: %v", err)
		// 即使获取失败，也返回成功，因为更新操作已经完成
		return &v1.UpdateCampaignReply{
			Campaign: &v1.Campaign{
				CampaignId: existing.CampaignID,
			},
		}, nil
	}

	// 将 biz.Campaign 转换为 v1.Campaign
	pbCampaign := &v1.Campaign{
		CampaignId:   updatedCampaign.CampaignID,
		CampaignName: updatedCampaign.CampaignName,
		TenantId:     updatedCampaign.TenantID,
		ProductCode:  updatedCampaign.ProductCode,
		CampaignType: updatedCampaign.CampaignType,
		StartTime:    updatedCampaign.StartTime.Format(time.RFC3339),
		EndTime:      updatedCampaign.EndTime.Format(time.RFC3339),
		Status:       int32(updatedCampaign.Status),
		CreatedAt:    updatedCampaign.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    updatedCampaign.UpdatedAt.Format(time.RFC3339),
	}

	return &v1.UpdateCampaignReply{
		Campaign: pbCampaign,
	}, nil
}

// GetCampaign 获取营销活动
func (s *MarketingService) GetCampaign(ctx context.Context, req *v1.GetCampaignRequest) (*v1.GetCampaignReply, error) {
	if req.CampaignId == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "campaign_id is required")
	}

	// 获取活动
	campaign, err := s.campaignUc.GetCampaign(ctx, req.CampaignId)
	if err != nil {
		return nil, status.Errorf(grpccodes.NotFound, "campaign not found: %v", err)
	}

	// 构建响应
	pbCampaign := &v1.Campaign{
		CampaignId:   campaign.CampaignID,
		TenantId:     campaign.TenantID,
		CampaignName: campaign.CampaignName,
		CampaignType: campaign.CampaignType,
		Status:       campaign.Status,
		StartTime:    campaign.StartTime.Format(time.RFC3339),
		EndTime:      campaign.EndTime.Format(time.RFC3339),
		CreatedAt:    campaign.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    campaign.UpdatedAt.Format(time.RFC3339),
	}

	// 添加规则配置
	// 在 proto 中，规则配置是作为 map<string, string> 存储的
	if len(campaign.Rules) > 0 {
		// 初始化 rule_config 映射
		pbCampaign.RuleConfig = make(map[string]string)

		// 将所有规则合并到一个映射中
		for _, rule := range campaign.Rules {
			// 将规则类型作为前缀，避免键冲突
			prefix := rule.RuleType + "_"

			// 将 map[string]interface{} 转换为 map[string]string
			for k, v := range rule.Config {
				// 将值转换为字符串
				var strValue string
				switch val := v.(type) {
				case string:
					strValue = val
				case float64, float32, int, int32, int64:
					strValue = fmt.Sprintf("%v", val)
				case bool:
					strValue = fmt.Sprintf("%v", val)
				default:
					// 对于复杂类型，尝试使用 JSON 序列化
					jsonBytes, err := json.Marshal(val)
					if err == nil {
						strValue = string(jsonBytes)
					} else {
						strValue = fmt.Sprintf("%v", val)
					}
				}

				// 添加到映射中，使用前缀避免冲突
				pbCampaign.RuleConfig[prefix+k] = strValue
			}
		}
	}

	return &v1.GetCampaignReply{
		Campaign: pbCampaign,
	}, nil
}

// ListCampaigns 列出营销活动
func (s *MarketingService) ListCampaigns(ctx context.Context, req *v1.ListCampaignsRequest) (*v1.ListCampaignsReply, error) {
	// 获取活动列表
	campaigns, total, err := s.campaignUc.ListCampaigns(ctx, req.TenantId, req.ProductCode, req.CampaignType, req.Status, req.PageNum, req.PageSize)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "list campaigns failed: %v", err)
	}

	// 构建响应
	pbCampaigns := make([]*v1.Campaign, 0, len(campaigns))
	for _, campaign := range campaigns {
		pbCampaign := &v1.Campaign{
			CampaignId:   campaign.CampaignID,
			TenantId:     campaign.TenantID,
			CampaignName: campaign.CampaignName,
			CampaignType: campaign.CampaignType,
			Status:       campaign.Status,
			StartTime:    campaign.StartTime.Format(time.RFC3339),
			EndTime:      campaign.EndTime.Format(time.RFC3339),
			CreatedAt:    campaign.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    campaign.UpdatedAt.Format(time.RFC3339),
		}

		// 添加规则配置
		// 在 proto 中，规则配置是作为 map<string, string> 存储的
		if len(campaign.Rules) > 0 {
			// 初始化 rule_config 映射
			pbCampaign.RuleConfig = make(map[string]string)

			// 将所有规则合并到一个映射中
			for _, rule := range campaign.Rules {
				// 将规则类型作为前缀，避免键冲突
				prefix := rule.RuleType + "_"

				// 将 map[string]interface{} 转换为 map[string]string
				for k, v := range rule.Config {
					// 将值转换为字符串
					var strValue string
					switch val := v.(type) {
					case string:
						strValue = val
					case float64, float32, int, int32, int64:
						strValue = fmt.Sprintf("%v", val)
					case bool:
						strValue = fmt.Sprintf("%v", val)
					default:
						// 对于复杂类型，尝试使用 JSON 序列化
						jsonBytes, err := json.Marshal(val)
						if err == nil {
							strValue = string(jsonBytes)
						} else {
							strValue = fmt.Sprintf("%v", val)
						}
					}

					// 添加到映射中，使用前缀避免冲突
					pbCampaign.RuleConfig[prefix+k] = strValue
				}
			}
		}

		pbCampaigns = append(pbCampaigns, pbCampaign)
	}

	return &v1.ListCampaignsReply{
		Campaigns: pbCampaigns,
		Total:     int32(total),
	}, nil
}

// DeleteCampaign 删除营销活动
func (s *MarketingService) DeleteCampaign(ctx context.Context, req *v1.DeleteCampaignRequest) (*v1.DeleteCampaignReply, error) {
	if req.CampaignId == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "campaign_id is required")
	}

	// 删除活动
	err := s.campaignUc.DeleteCampaign(ctx, req.CampaignId)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "delete campaign failed: %v", err)
	}

	return &v1.DeleteCampaignReply{
		Success: true,
	}, nil
}

// GenerateRedeemCodes 生成兑换码
func (s *MarketingService) GenerateRedeemCodes(ctx context.Context, req *v1.GenerateRedeemCodesRequest) (*v1.GenerateRedeemCodesReply, error) {
	if req.CampaignId == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "campaign_id is required")
	}
	if req.CodeType == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "code_type is required")
	}
	if req.GenerateType == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "generate_type is required")
	}
	if req.Count <= 0 {
		return nil, status.Error(grpccodes.InvalidArgument, "count must be greater than 0")
	}

	// 解析过期时间
	var expireTime time.Time
	var err error
	if req.ExpireTime != "" {
		expireTime, err = time.Parse(time.RFC3339, req.ExpireTime)
		if err != nil {
			return nil, status.Errorf(grpccodes.InvalidArgument, "invalid expire_time format: %v", err)
		}
	} else {
		// 默认过期时间为一年后
		expireTime = time.Now().AddDate(1, 0, 0)
	}

	// 生成兑换码
	codeBatch, sampleCodes, err := s.redeemCodeUc.GenerateRedeemCodes(
		ctx,
		req.CampaignId,
		req.CodeType,
		req.GenerateType,
		req.Count,
		req.GenerateRule,
		float64(req.ValueAmount),
		float64(req.ValuePercent),
		req.RewardItems,
		expireTime,
		req.Operator,
	)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "generate redeem codes failed: %v", err)
	}

	// 构建响应
	return &v1.GenerateRedeemCodesReply{
		BatchId:     codeBatch.BatchID,
		TotalCount:  codeBatch.TotalCount,
		SampleCodes: sampleCodes,
	}, nil
}

// GetRedeemCode 获取兑换码
func (s *MarketingService) GetRedeemCode(ctx context.Context, req *v1.GetRedeemCodeRequest) (*v1.GetRedeemCodeReply, error) {
	if req.Code == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "code is required")
	}

	// 获取兑换码
	code, err := s.redeemCodeUc.GetRedeemCode(ctx, req.Code)
	if err != nil {
		return nil, status.Errorf(grpccodes.NotFound, "redeem code not found: %v", err)
	}

	// 构建响应
	pbCode := &v1.RedeemCode{
		Code:        code.Code,
		CampaignId:  code.CampaignID,
		TenantId:    code.TenantID,
		ProductCode: code.ProductCode,
		CodeType:    code.CodeType,
		UserId:      code.UserID,
		Status:      code.Status,
		ValidUntil:  timestamppb.New(code.ValidUntil),
		CreatedAt:   timestamppb.New(code.CreatedAt),
		UpdatedAt:   timestamppb.New(code.UpdatedAt),
	}

	if code.RedemptionAt != nil {
		pbCode.RedemptionAt = timestamppb.New(*code.RedemptionAt)
	}

	return &v1.GetRedeemCodeReply{
		Code: pbCode,
	}, nil
}

// RedeemCode 兑换码兑换
func (s *MarketingService) RedeemCode(ctx context.Context, req *v1.RedeemCodeRequest) (*v1.RedeemCodeReply, error) {
	if req.Code == "" {
		return nil, status.Error(grpccodes.InvalidArgument, "code is required")
	}

	// 获取用户 ID
	var userID int64 = req.UserId

	// 兑换码兑换
	// 调用 biz 层的 RedeemCode 方法，传递所有必要的参数
	code, _, err := s.redeemCodeUc.RedeemCode(
		ctx,
		req.Code,
		userID,
		req.ProductCode,
		req.RedeemChannel,
		req.DeviceInfo,
		req.IpAddress, // 使用正确的字段名称
		req.Location,
		req.OrderId,
	)

	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "redeem code failed: %v", err)
	}

	// 构建响应
	// 将 biz.RedeemCode 转换为 v1.RedeemCode
	now := time.Now()
	validUntil := now.Add(24 * time.Hour) // 默认24小时后过期

	pbCode := &v1.RedeemCode{
		Code:          code.Code,
		CampaignId:    code.CampaignID,
		TenantId:      code.TenantID,
		ProductCode:   code.ProductCode,
		CodeType:      code.CodeType,
		UserId:        code.UserID,
		Status:        int32(code.Status),
		ValueAmount:   float32(code.ValueAmount),
		ValuePercent:  float32(code.ValuePercent),
		RewardItems:   make(map[string]string),
		ValidFrom:     timestamppb.New(now),
		ValidUntil:    timestamppb.New(validUntil),
		RedemptionAt:  timestamppb.New(now),
		RedeemChannel: code.RedeemChannel,
		CreatedAt:     timestamppb.New(code.CreatedAt),
		UpdatedAt:     timestamppb.New(code.UpdatedAt),
	}

	return &v1.RedeemCodeReply{
		Success:      true,
		Message:      "Redemption successful",
		RewardDetail: make(map[string]string),
		CodeInfo:     pbCode,
	}, nil
}

// ListRedeemCodes 列出兑换码
func (s *MarketingService) ListRedeemCodes(ctx context.Context, req *v1.ListRedeemCodesRequest) (*v1.ListRedeemCodesReply, error) {
	// 获取兑换码列表
	// 根据 redeemCodeUc.ListRedeemCodes 的方法签名传递所有必要的参数
	codes, total, err := s.redeemCodeUc.ListRedeemCodes(
		ctx,
		req.CampaignId,
		req.TenantId,
		req.ProductCode,
		req.CodeType,
		req.Status,
		req.UserId,
		req.PageNum,
		req.PageSize,
	)
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "list redeem codes failed: %v", err)
	}

	// 构建响应
	pbCodes := make([]*v1.RedeemCode, 0, len(codes))
	for _, code := range codes {
		pbCode := &v1.RedeemCode{
			Code:        code.Code,
			CampaignId:  code.CampaignID,
			TenantId:    code.TenantID,
			ProductCode: code.ProductCode,
			CodeType:    code.CodeType,
			UserId:      code.UserID,
			Status:      code.Status,
			ValidUntil:  timestamppb.New(code.ValidUntil),
			CreatedAt:   timestamppb.New(code.CreatedAt),
			UpdatedAt:   timestamppb.New(code.UpdatedAt),
		}

		if code.RedemptionAt != nil {
			pbCode.RedemptionAt = timestamppb.New(*code.RedemptionAt)
		}

		pbCodes = append(pbCodes, pbCode)
	}

	return &v1.ListRedeemCodesReply{
		Codes: pbCodes,
		Total: int32(total),
	}, nil
}

/*
// ListCodeBatches 列出兑换码批次
// 注意：在 proto 文件中没有定义 ListCodeBatches 相关的消息类型和方法，暂时注释掉

func (s *MarketingService) ListCodeBatches(ctx context.Context, req *v1.ListCodeBatchesRequest) (*v1.ListCodeBatchesReply, error) {
	// 获取批次列表
	batches, total, err := s.redeemCodeUc.ListCodeBatches(ctx, req.CampaignId, int(req.Page-1)*int(req.PageSize), int(req.PageSize))
	if err != nil {
		return nil, status.Errorf(grpccodes.Internal, "list code batches failed: %v", err)
	}

	// 构建响应
	pbBatches := make([]*v1.CodeBatch, 0, len(batches))
	for _, batch := range batches {
		// 将生成规则转换为 map
		generateRule := make(map[string]string)
		// 添加批次相关信息到生成规则中，因为 proto 中没有对应字段
		generateRule["batch_name"] = batch.BatchName
		generateRule["code_count"] = fmt.Sprintf("%d", batch.CodeCount)
		generateRule["code_prefix"] = batch.CodePrefix
		generateRule["code_type"] = batch.CodeType
		generateRule["code_length"] = fmt.Sprintf("%d", batch.CodeLength)
		generateRule["valid_from"] = batch.ValidFrom.Format(time.RFC3339)
		generateRule["valid_until"] = batch.ValidUntil.Format(time.RFC3339)
		generateRule["description"] = batch.Description

		pbBatches = append(pbBatches, &v1.CodeBatch{
			BatchId:      batch.BatchID,
			CampaignId:   batch.CampaignID,
			TenantId:     batch.TenantID,
			GenerateType: batch.GenerateType,
			TotalCount:   batch.TotalCount,
			UsedCount:    batch.UsedCount,
			Operator:     batch.Operator,
			GenerateRule: generateRule,
			CreatedAt:    batch.CreatedAt.Format(time.RFC3339),
		})
	}

	return &v1.ListCodeBatchesReply{
		Batches: pbBatches,
		Total:   int32(total),
	}, nil
}
*/
