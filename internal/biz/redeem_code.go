package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"marketing-service/internal/constants"
)

// RedeemCode 兑换码领域对象
type RedeemCode struct {
	Code         string
	TenantID     string
	AppID        string
	GrantID      string
	CampaignID   string
	CampaignName string
	RewardID     string
	RewardName   string
	BatchID      string
	Status       string
	OwnerUserID  *int64
	RedeemedBy   *int64
	RedeemedAt   *time.Time
	ExpireAt     *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RedeemCodeRepo 兑换码仓储接口
type RedeemCodeRepo interface {
	Save(context.Context, *RedeemCode) (*RedeemCode, error)
	FindByCode(context.Context, string, string) (*RedeemCode, error)
	List(context.Context, string, string, string, string, int64, string, int, int) ([]*RedeemCode, int64, error)
	UpdateStatus(context.Context, string, string, string) error
	Redeem(context.Context, string, string, int64) error
	BatchCreate(context.Context, []*RedeemCode) error
}

// RedeemCodeUseCase 兑换码用例
type RedeemCodeUseCase struct {
	repo RedeemCodeRepo
	log  *log.Helper
}

// NewRedeemCodeUseCase 创建兑换码用例
func NewRedeemCodeUseCase(repo RedeemCodeRepo, logger log.Logger) *RedeemCodeUseCase {
	return &RedeemCodeUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建兑换码
func (uc *RedeemCodeUseCase) Create(ctx context.Context, rc *RedeemCode) (*RedeemCode, error) {
	if rc.Status == "" {
		rc.Status = constants.RedeemCodeStatusActive
	}
	rc.CreatedAt = time.Now()
	rc.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, rc)
}

// GetByCode 根据兑换码获取
func (uc *RedeemCodeUseCase) GetByCode(ctx context.Context, code, tenantID string) (*RedeemCode, error) {
	return uc.repo.FindByCode(ctx, code, tenantID)
}

// List 列出兑换码
func (uc *RedeemCodeUseCase) List(ctx context.Context, tenantID, appID, campaignID, batchID string, userID int64, status string, page, pageSize int) ([]*RedeemCode, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, campaignID, batchID, userID, status, page, pageSize)
}

// Redeem 兑换码核销
func (uc *RedeemCodeUseCase) Redeem(ctx context.Context, code, tenantID string, userID int64) error {
	return uc.repo.Redeem(ctx, code, tenantID, userID)
}

// BatchCreate 批量创建兑换码
func (uc *RedeemCodeUseCase) BatchCreate(ctx context.Context, codes []*RedeemCode) error {
	return uc.repo.BatchCreate(ctx, codes)
}

// UpdateStatus 更新状态
func (uc *RedeemCodeUseCase) UpdateStatus(ctx context.Context, code, tenantID, status string) error {
	return uc.repo.UpdateStatus(ctx, code, tenantID, status)
}

