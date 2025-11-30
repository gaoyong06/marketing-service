package data_test

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"marketing-service/internal/biz"
	"marketing-service/internal/data"
	"marketing-service/internal/data/model"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(&model.Campaign{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	cleanup := func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}

	return db, cleanup
}

func TestCampaignRepo_Save(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := log.NewStdLogger(nil)
	dataData, _, _ := data.NewData(db, nil, logger)
	repo := data.NewCampaignRepo(dataData, nil, logger)

	ctx := context.Background()
	campaign := &biz.Campaign{
		ID:        "campaign-1",
		TenantID:  "tenant1",
		AppID:     "app1",
		Name:      "Test Campaign",
		Type:      "REDEEM_CODE",
		Status:    "ACTIVE",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := repo.Save(ctx, campaign)
	assert.NoError(t, err)
	assert.Equal(t, campaign.ID, result.ID)
	assert.Equal(t, campaign.Name, result.Name)
}

func TestCampaignRepo_FindByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := log.NewStdLogger(nil)
	dataData, _, _ := data.NewData(db, nil, logger)
	repo := data.NewCampaignRepo(dataData, nil, logger)

	ctx := context.Background()

	// 先创建
	campaign := &biz.Campaign{
		ID:        "campaign-1",
		TenantID:  "tenant1",
		AppID:     "app1",
		Name:      "Test Campaign",
		Type:      "REDEEM_CODE",
		Status:    "ACTIVE",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := repo.Save(ctx, campaign)
	assert.NoError(t, err)

	// 再查询
	result, err := repo.FindByID(ctx, "campaign-1")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, campaign.ID, result.ID)
	assert.Equal(t, campaign.Name, result.Name)
}

func TestCampaignRepo_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := log.NewStdLogger(nil)
	dataData, _, _ := data.NewData(db, nil, logger)
	repo := data.NewCampaignRepo(dataData, nil, logger)

	ctx := context.Background()

	// 创建多个活动
	for i := 1; i <= 3; i++ {
		campaign := &biz.Campaign{
			ID:        "campaign-" + string(rune('0'+i)),
			TenantID:  "tenant1",
			AppID:     "app1",
			Name:      "Test Campaign " + string(rune('0'+i)),
			Type:      "REDEEM_CODE",
			Status:    "ACTIVE",
			StartTime: time.Now(),
			EndTime:   time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_, err := repo.Save(ctx, campaign)
		assert.NoError(t, err)
	}

	// 查询列表
	result, total, err := repo.List(ctx, "tenant1", "app1", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, result, 3)
}

