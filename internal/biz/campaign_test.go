package biz_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"marketing-service/internal/biz"
)

// MockCampaignRepo 是 CampaignRepo 的 mock 实现
type MockCampaignRepo struct {
	mock.Mock
}

func (m *MockCampaignRepo) Save(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*biz.Campaign), args.Error(1)
}

func (m *MockCampaignRepo) Update(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	args := m.Called(ctx, c)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*biz.Campaign), args.Error(1)
}

func (m *MockCampaignRepo) FindByID(ctx context.Context, id string) (*biz.Campaign, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*biz.Campaign), args.Error(1)
}

func (m *MockCampaignRepo) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*biz.Campaign, int64, error) {
	args := m.Called(ctx, tenantID, appID, page, pageSize)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*biz.Campaign), args.Get(1).(int64), args.Error(2)
}

func (m *MockCampaignRepo) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestCampaignUseCase_Create(t *testing.T) {
	logger := log.NewStdLogger(nil)
	repo := new(MockCampaignRepo)
	uc := biz.NewCampaignUseCase(repo, logger)

	ctx := context.Background()
	campaign := &biz.Campaign{
		TenantID: "tenant1",
		AppID:    "app1",
		Name:     "Test Campaign",
		Type:     "REDEEM_CODE",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
	}

	// Mock 返回的 campaign 应该包含生成的 ID
	expected := &biz.Campaign{
		ID:        "campaign-1",
		TenantID:  "tenant1",
		AppID:     "app1",
		Name:      "Test Campaign",
		Type:      "REDEEM_CODE",
		Status:    "ACTIVE",
		StartTime: campaign.StartTime,
		EndTime:   campaign.EndTime,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("Save", ctx, mock.MatchedBy(func(c *biz.Campaign) bool {
		return c.TenantID == "tenant1" && c.Name == "Test Campaign" && c.ID != ""
	})).Return(expected, nil)

	result, err := uc.Create(ctx, campaign)
	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "ACTIVE", result.Status)
	repo.AssertExpectations(t)
}

func TestCampaignUseCase_Get(t *testing.T) {
	logger := log.NewStdLogger(nil)
	repo := new(MockCampaignRepo)
	uc := biz.NewCampaignUseCase(repo, logger)

	ctx := context.Background()
	expected := &biz.Campaign{
		ID:        "campaign-1",
		TenantID:  "tenant1",
		AppID:     "app1",
		Name:      "Test Campaign",
		Type:      "REDEEM_CODE",
		Status:    "ACTIVE",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	repo.On("FindByID", ctx, "campaign-1").Return(expected, nil)

	result, err := uc.Get(ctx, "campaign-1")
	assert.NoError(t, err)
	assert.Equal(t, expected.ID, result.ID)
	assert.Equal(t, expected.Name, result.Name)
	repo.AssertExpectations(t)
}

func TestCampaignUseCase_List(t *testing.T) {
	logger := log.NewStdLogger(nil)
	repo := new(MockCampaignRepo)
	uc := biz.NewCampaignUseCase(repo, logger)

	ctx := context.Background()
	expected := []*biz.Campaign{
		{
			ID:        "campaign-1",
			TenantID:  "tenant1",
			AppID:     "app1",
			Name:      "Test Campaign 1",
			Type:      "REDEEM_CODE",
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
		},
		{
			ID:        "campaign-2",
			TenantID:  "tenant1",
			AppID:     "app1",
			Name:      "Test Campaign 2",
			Type:      "TASK_REWARD",
			Status:    "ACTIVE",
			CreatedAt: time.Now(),
		},
	}

	repo.On("List", ctx, "tenant1", "app1", 1, 20).Return(expected, int64(2), nil)

	result, total, err := uc.List(ctx, "tenant1", "app1", 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

