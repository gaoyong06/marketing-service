package biz

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
)

// TaskCompletionLog 任务完成记录领域对象
type TaskCompletionLog struct {
	CompletionID string
	TaskID       string
	TaskName     string
	CampaignID   string
	CampaignName string
	UserID       int64
	TenantID     string
	AppID        string
	GrantID      string
	ProgressData string // JSON string
	TriggerEvent string
	CompletedAt  time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TaskCompletionLogRepo 任务完成记录仓储接口
type TaskCompletionLogRepo interface {
	Save(context.Context, *TaskCompletionLog) (*TaskCompletionLog, error)
	FindByID(context.Context, string) (*TaskCompletionLog, error)
	List(context.Context, string, string, string, string, int64, int, int) ([]*TaskCompletionLog, int64, error)
	CountByTaskAndUser(context.Context, string, int64) (int64, error)
	CountByTask(context.Context, string, string) (int64, error)              // 统计任务总完成次数
	CountUniqueUsersByTask(context.Context, string, string) (int64, error)   // 统计任务的唯一用户数
}

// TaskCompletionLogUseCase 任务完成记录用例
type TaskCompletionLogUseCase struct {
	repo TaskCompletionLogRepo
	log  *log.Helper
}

// NewTaskCompletionLogUseCase 创建任务完成记录用例
func NewTaskCompletionLogUseCase(repo TaskCompletionLogRepo, logger log.Logger) *TaskCompletionLogUseCase {
	return &TaskCompletionLogUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建任务完成记录
func (uc *TaskCompletionLogUseCase) Create(ctx context.Context, log *TaskCompletionLog) (*TaskCompletionLog, error) {
	if log.CompletionID == "" {
		log.CompletionID = uuid.New().String()
	}
	if log.CompletedAt.IsZero() {
		log.CompletedAt = time.Now()
	}
	log.CreatedAt = time.Now()
	log.UpdatedAt = time.Now()
	return uc.repo.Save(ctx, log)
}

// Get 获取任务完成记录
func (uc *TaskCompletionLogUseCase) Get(ctx context.Context, id string) (*TaskCompletionLog, error) {
	return uc.repo.FindByID(ctx, id)
}

// List 列出任务完成记录
func (uc *TaskCompletionLogUseCase) List(ctx context.Context, tenantID, appID, taskID, campaignID string, userID int64, page, pageSize int) ([]*TaskCompletionLog, int64, error) {
	return uc.repo.List(ctx, tenantID, appID, taskID, campaignID, userID, page, pageSize)
}

// CountByTaskAndUser 统计用户完成某个任务的次数
func (uc *TaskCompletionLogUseCase) CountByTaskAndUser(ctx context.Context, taskID string, userID int64) (int64, error) {
	return uc.repo.CountByTaskAndUser(ctx, taskID, userID)
}

// CountByTask 统计任务总完成次数
func (uc *TaskCompletionLogUseCase) CountByTask(ctx context.Context, taskID, campaignID string) (int64, error) {
	return uc.repo.CountByTask(ctx, taskID, campaignID)
}

// CountUniqueUsersByTask 统计任务的唯一用户数
func (uc *TaskCompletionLogUseCase) CountUniqueUsersByTask(ctx context.Context, taskID, campaignID string) (int64, error) {
	return uc.repo.CountUniqueUsersByTask(ctx, taskID, campaignID)
}

