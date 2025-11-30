package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// CampaignTask 活动-任务关联领域对象
type CampaignTask struct {
	CampaignTaskID int64
	CampaignID     string
	TaskID         string
	Config         string // JSON string
	SortOrder      int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// CampaignTaskRepo 活动-任务关联仓储接口
type CampaignTaskRepo interface {
	Save(context.Context, *CampaignTask) (*CampaignTask, error)
	Delete(context.Context, string, string) error
	ListByCampaign(context.Context, string) ([]*CampaignTask, error)
}

// CampaignTaskUseCase 活动-任务关联用例
type CampaignTaskUseCase struct {
	repo CampaignTaskRepo
	log  *log.Helper
}

// NewCampaignTaskUseCase 创建活动-任务关联用例
func NewCampaignTaskUseCase(repo CampaignTaskRepo, logger log.Logger) *CampaignTaskUseCase {
	return &CampaignTaskUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// AddTaskToCampaign 将任务添加到活动
func (uc *CampaignTaskUseCase) AddTaskToCampaign(ctx context.Context, campaignID, taskID string, sortOrder int, config string) (*CampaignTask, error) {
	ct := &CampaignTask{
		CampaignID: campaignID,
		TaskID:     taskID,
		SortOrder:  sortOrder,
		Config:     config,
	}
	return uc.repo.Save(ctx, ct)
}

// RemoveTaskFromCampaign 从活动中移除任务
func (uc *CampaignTaskUseCase) RemoveTaskFromCampaign(ctx context.Context, campaignID, taskID string) error {
	return uc.repo.Delete(ctx, campaignID, taskID)
}

// ListCampaignTasks 列出活动的所有任务
func (uc *CampaignTaskUseCase) ListCampaignTasks(ctx context.Context, campaignID string) ([]*CampaignTask, error) {
	return uc.repo.ListByCampaign(ctx, campaignID)
}

