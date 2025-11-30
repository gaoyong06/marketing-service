package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Reward 奖励领域对象
type Reward struct {
	ID                string
	TenantID          string
	AppID             string
	RewardType        string
	Name              string
	ContentConfig     string // JSON string
	GeneratorConfig   string // JSON string
	DistributorConfig string // JSON string
	ValidatorConfig   string // JSON string
	Version           int
	ValidDays         int
	ExtraConfig       string // JSON string
	Status            string
	Description       string
	CreatedBy         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// RewardRepo 奖励仓储接口
type RewardRepo interface {
	Save(context.Context, *Reward) (*Reward, error)
	Update(context.Context, *Reward) (*Reward, error)
	FindByID(context.Context, string) (*Reward, error)
	List(context.Context, string, string, int, int) ([]*Reward, int64, error)
	Delete(context.Context, string) error
}

// RewardUseCase 奖励用例
type RewardUseCase struct {
	repo RewardRepo
	log  *log.Helper
}

// NewRewardUseCase 创建奖励用例
func NewRewardUseCase(repo RewardRepo, logger log.Logger) *RewardUseCase {
	return &RewardUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建奖励
func (uc *RewardUseCase) Create(ctx context.Context, r *Reward) (*Reward, error) {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	if r.Status == "" {
		r.Status = "ACTIVE"
	}
	if r.Version == 0 {
		r.Version = 1
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, r)
}

// Update 更新奖励（版本号自动递增）
func (uc *RewardUseCase) Update(ctx context.Context, r *Reward) (*Reward, error) {
	r.UpdatedAt = time.Now()
	// 版本号在 Repository 层自动递增
	return uc.repo.Update(ctx, r)
}

// Get 获取奖励
func (uc *RewardUseCase) Get(ctx context.Context, id string) (*Reward, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 列出奖励
func (uc *RewardUseCase) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*Reward, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, page, pageSize)
}

// Delete 删除奖励
func (uc *RewardUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}

