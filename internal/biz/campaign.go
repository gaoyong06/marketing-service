package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// Campaign is a Campaign domain object.
type Campaign struct {
	ID              string
	TenantID        string
	AppID           string
	Name            string
	Type            string
	StartTime       time.Time
	EndTime         time.Time
	AudienceConfig  string // JSON string
	ValidatorConfig string // JSON string
	Status          string
	Description     string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// CampaignRepo is a Greater repo.
type CampaignRepo interface {
	Save(context.Context, *Campaign) (*Campaign, error)
	Update(context.Context, *Campaign) (*Campaign, error)
	FindByID(context.Context, string) (*Campaign, error)
	List(context.Context, string, string, int, int) ([]*Campaign, int64, error)
	Delete(context.Context, string) error
}

// CampaignUseCase is a Campaign usecase.
type CampaignUseCase struct {
	repo CampaignRepo
	log  *log.Helper
}

// NewCampaignUseCase new a Campaign usecase.
func NewCampaignUseCase(repo CampaignRepo, logger log.Logger) *CampaignUseCase {
	return &CampaignUseCase{repo: repo, log: log.NewHelper(logger)}
}

// Create creates a Campaign, and returns the new Campaign.
func (uc *CampaignUseCase) Create(ctx context.Context, c *Campaign) (*Campaign, error) {
	if c.ID == "" {
		c.ID = uuid.New().String()
	}
	if c.Status == "" {
		c.Status = "ACTIVE"
	}
	c.CreatedAt = time.Now()
	c.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, c)
}

// Update updates a Campaign.
func (uc *CampaignUseCase) Update(ctx context.Context, c *Campaign) (*Campaign, error) {
	c.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, c)
}

// Get gets a Campaign.
func (uc *CampaignUseCase) Get(ctx context.Context, id string) (*Campaign, error) {
	return uc.repo.FindByID(ctx, id)
}

// List lists Campaigns.
func (uc *CampaignUseCase) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*Campaign, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, page, pageSize)
}

// Delete deletes a Campaign.
func (uc *CampaignUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
