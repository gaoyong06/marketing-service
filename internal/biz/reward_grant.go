package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// RewardGrant 奖励发放领域对象
type RewardGrant struct {
	GrantID         string
	RewardID        string
	RewardName      string
	RewardType      string
	RewardVersion   int
	ContentSnapshot string // JSON string
	GeneratorConfig string // JSON string
	CampaignID      string
	CampaignName    string
	TaskID          string
	TaskName        string
	TenantID        string
	AppID           string
	UserID          int64
	Status          string
	ReservedAt      *time.Time
	DistributedAt   *time.Time
	UsedAt          *time.Time
	ExpireTime      *time.Time
	ErrorMessage    string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// RewardGrantRepo 奖励发放仓储接口
type RewardGrantRepo interface {
	Save(context.Context, *RewardGrant) (*RewardGrant, error)
	Update(context.Context, *RewardGrant) (*RewardGrant, error)
	FindByID(context.Context, string) (*RewardGrant, error)
	List(context.Context, string, string, int64, string, int, int) ([]*RewardGrant, int64, error)
	UpdateStatus(context.Context, string, string) error
	CountByStatus(context.Context, string, string) (int64, error)
	BatchSave(context.Context, []*RewardGrant) error // 批量保存（性能优化）
}

// RewardGrantUseCase 奖励发放用例
type RewardGrantUseCase struct {
	repo RewardGrantRepo
	log  *log.Helper
}

// NewRewardGrantUseCase 创建奖励发放用例
func NewRewardGrantUseCase(repo RewardGrantRepo, logger log.Logger) *RewardGrantUseCase {
	return &RewardGrantUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建奖励发放记录
func (uc *RewardGrantUseCase) Create(ctx context.Context, grant *RewardGrant) (*RewardGrant, error) {
	if grant.GrantID == "" {
		grant.GrantID = uuid.New().String()
	}
	if grant.Status == "" {
		grant.Status = "PENDING"
	}
	grant.CreatedAt = time.Now()
	grant.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, grant)
}

// Update 更新奖励发放记录
func (uc *RewardGrantUseCase) Update(ctx context.Context, grant *RewardGrant) (*RewardGrant, error) {
	grant.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, grant)
}

// Get 获取奖励发放记录
func (uc *RewardGrantUseCase) Get(ctx context.Context, id string) (*RewardGrant, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 列出奖励发放记录
func (uc *RewardGrantUseCase) List(ctx context.Context, tenantID, appID string, userID int64, status string, page, pageSize int) ([]*RewardGrant, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, userID, status, page, pageSize)
}

// UpdateStatus 更新状态
func (uc *RewardGrantUseCase) UpdateStatus(ctx context.Context, grantID, status string) error {
	return uc.repo.UpdateStatus(ctx, grantID, status)
}

// CountByStatus 按状态统计数量
func (uc *RewardGrantUseCase) CountByStatus(ctx context.Context, rewardID, status string) (int64, error) {
	return uc.repo.CountByStatus(ctx, rewardID, status)
}

// BatchCreate 批量创建奖励发放记录（性能优化）
func (uc *RewardGrantUseCase) BatchCreate(ctx context.Context, grants []*RewardGrant) error {
	if len(grants) == 0 {
		return nil
	}

	// 设置默认值
	now := time.Now()
	for _, grant := range grants {
		if grant.GrantID == "" {
			grant.GrantID = uuid.New().String()
		}
		if grant.Status == "" {
			grant.Status = "GENERATED"
		}
		if grant.CreatedAt.IsZero() {
			grant.CreatedAt = now
		}
		if grant.UpdatedAt.IsZero() {
			grant.UpdatedAt = now
		}
	}

	// 批量保存
	return uc.repo.BatchSave(ctx, grants)
}
