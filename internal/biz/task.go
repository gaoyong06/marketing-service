package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"marketing-service/internal/constants"
)

// Task 任务领域对象
type Task struct {
	ID              string
	TenantID        string
	AppID           string
	Name            string
	TaskType        string
	TriggerConfig   string // JSON string
	ConditionConfig string // JSON string
	RewardID        string
	Status          string
	StartTime       time.Time
	EndTime         time.Time
	MaxCount        int
	Description     string
	CreatedBy       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TaskRepo 任务仓储接口
type TaskRepo interface {
	Save(context.Context, *Task) (*Task, error)
	Update(context.Context, *Task) (*Task, error)
	FindByID(context.Context, string) (*Task, error)
	List(context.Context, string, string, int, int) ([]*Task, int64, error)
	ListByCampaign(context.Context, string) ([]*Task, error)
	ListActive(context.Context, string, string) ([]*Task, error)
	Delete(context.Context, string) error
}

// TaskUseCase 任务用例
type TaskUseCase struct {
	repo TaskRepo
	log  *log.Helper
}

// NewTaskUseCase 创建任务用例
func NewTaskUseCase(repo TaskRepo, logger log.Logger) *TaskUseCase {
	return &TaskUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建任务
func (uc *TaskUseCase) Create(ctx context.Context, t *Task) (*Task, error) {
	if t.ID == "" {
		t.ID = GenerateShortID()
	}
	if t.Status == "" {
		t.Status = constants.StatusActive
	}
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, t)
}

// Update 更新任务
func (uc *TaskUseCase) Update(ctx context.Context, t *Task) (*Task, error) {
	t.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, t)
}

// Get 获取任务
func (uc *TaskUseCase) Get(ctx context.Context, id string) (*Task, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 列出任务
func (uc *TaskUseCase) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*Task, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, page, pageSize)
}

// ListByCampaign 根据活动列出任务
func (uc *TaskUseCase) ListByCampaign(ctx context.Context, campaignID string) ([]*Task, error) {
	return uc.repo.ListByCampaign(ctx, campaignID)
}

// ListActive 列出活跃任务
func (uc *TaskUseCase) ListActive(ctx context.Context, tenantID, appID string) ([]*Task, error) {
	return uc.repo.ListActive(ctx, tenantID, appID)
}

// Delete 删除任务
func (uc *TaskUseCase) Delete(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}
