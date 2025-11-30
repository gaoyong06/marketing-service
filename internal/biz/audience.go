package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Audience 受众领域对象
type Audience struct {
	ID           string
	TenantID     string
	AppID        string
	Name         string
	AudienceType string
	RuleConfig   string // JSON string
	Status       string
	Description  string
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AudienceRepo 受众仓储接口
type AudienceRepo interface {
	Save(context.Context, *Audience) (*Audience, error)
	Update(context.Context, *Audience) (*Audience, error)
	FindByID(context.Context, string) (*Audience, error)
	List(context.Context, string, string, int, int) ([]*Audience, int64, error)
	Delete(context.Context, string) error
}

// AudienceUseCase 受众用例
type AudienceUseCase struct {
	repo AudienceRepo
	log  *log.Helper
}

// NewAudienceUseCase 创建受众用例
func NewAudienceUseCase(repo AudienceRepo, logger log.Logger) *AudienceUseCase {
	return &AudienceUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建受众
func (uc *AudienceUseCase) Create(ctx context.Context, a *Audience) (*Audience, error) {
	if a.ID == "" {
		a.ID = GenerateShortID()
	}
	if a.Status == "" {
		a.Status = "ACTIVE"
	}
	a.CreatedAt = time.Now()
	a.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, a)
}

// Update 更新受众
func (uc *AudienceUseCase) Update(ctx context.Context, a *Audience) (*Audience, error) {
	a.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, a)
}

// Get 获取受众
func (uc *AudienceUseCase) Get(ctx context.Context, id string) (*Audience, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 列出受众
func (uc *AudienceUseCase) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*Audience, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, page, pageSize)
}

// Delete 删除受众
func (uc *AudienceUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
