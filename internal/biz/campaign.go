package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// CampaignRule 活动规则
type CampaignRule struct {
	RuleType string                 // 规则类型
	Config   map[string]interface{} // 规则配置
}

// Campaign 营销活动领域模型
type Campaign struct {
	CampaignID   string          // 活动ID
	CampaignName string          // 活动名称
	TenantID     string          // 所属租户
	ProductCode  string          // 适用产品线
	CampaignType string          // 活动类型
	StartTime    time.Time       // 开始时间
	EndTime      time.Time       // 结束时间
	TotalBudget  float64         // 总预算
	Rules        []*CampaignRule // 活动规则
	Description  string          // 活动描述
	Status       int32           // 状态：0-未开始 1-进行中 2-已结束 3-手动终止
	CreatedBy    string          // 创建人
	CreatedAt    time.Time       // 创建时间
	UpdatedAt    time.Time       // 更新时间
}

// CampaignRepo 活动仓储接口
type CampaignRepo interface {
	Create(ctx context.Context, campaign *Campaign) (*Campaign, error)
	Get(ctx context.Context, id string) (*Campaign, error)
	Update(ctx context.Context, campaign *Campaign) (*Campaign, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, tenantID, productCode, campaignType string, status int32, pageNum, pageSize int32) ([]*Campaign, int32, error)
}

// CampaignUsecase 活动用例
type CampaignUsecase struct {
	repo       CampaignRepo
	redeemRepo RedeemCodeRepo
	tenantRepo TenantRepo
	log        *log.Helper
}

// NewCampaignUsecase 创建活动用例
func NewCampaignUsecase(repo CampaignRepo, redeemRepo RedeemCodeRepo, tenantRepo TenantRepo, logger log.Logger) *CampaignUsecase {
	return &CampaignUsecase{
		repo:       repo,
		redeemRepo: redeemRepo,
		tenantRepo: tenantRepo,
		log:        log.NewHelper(logger),
	}
}

// CreateCampaign 创建活动
func (uc *CampaignUsecase) CreateCampaign(ctx context.Context, campaign *Campaign) (*Campaign, error) {
	uc.log.WithContext(ctx).Infof("CreateCampaign: %v", campaign.CampaignName)

	// 检查租户配额
	success, _, err := uc.tenantRepo.CheckAndConsumeQuota(ctx, campaign.TenantID, "MARKETING_CAMPAIGN", "TOTAL", 1, campaign.ProductCode)
	if err != nil || !success {
		return nil, err
	}

	return uc.repo.Create(ctx, campaign)
}

// GetCampaign 获取活动
func (uc *CampaignUsecase) GetCampaign(ctx context.Context, id string) (*Campaign, error) {
	uc.log.WithContext(ctx).Infof("GetCampaign: %v", id)
	return uc.repo.Get(ctx, id)
}

// UpdateCampaign 更新活动
func (uc *CampaignUsecase) UpdateCampaign(ctx context.Context, campaign *Campaign) (*Campaign, error) {
	uc.log.WithContext(ctx).Infof("UpdateCampaign: %v", campaign.CampaignID)
	return uc.repo.Update(ctx, campaign)
}

// DeleteCampaign 删除活动
func (uc *CampaignUsecase) DeleteCampaign(ctx context.Context, id string) error {
	uc.log.WithContext(ctx).Infof("DeleteCampaign: %v", id)
	return uc.repo.Delete(ctx, id)
}

// ListCampaigns 列出活动
func (uc *CampaignUsecase) ListCampaigns(ctx context.Context, tenantID, productCode, campaignType string, status int32, pageNum, pageSize int32) ([]*Campaign, int32, error) {
	uc.log.WithContext(ctx).Infof("ListCampaigns: tenantID=%v, productCode=%v", tenantID, productCode)
	return uc.repo.List(ctx, tenantID, productCode, campaignType, status, pageNum, pageSize)
}
