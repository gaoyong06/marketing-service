package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// taskCompletionLogRepo 实现 biz.TaskCompletionLogRepo 接口
type taskCompletionLogRepo struct {
	data *Data
	log  *log.Helper
}

// NewTaskCompletionLogRepo 创建 TaskCompletionLog Repository
func NewTaskCompletionLogRepo(data *Data, logger log.Logger) biz.TaskCompletionLogRepo {
	return &taskCompletionLogRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *taskCompletionLogRepo) toBizModel(m *model.TaskCompletionLog) *biz.TaskCompletionLog {
	if m == nil {
		return nil
	}
	return &biz.TaskCompletionLog{
		CompletionID: m.CompletionID,
		TaskID:      m.TaskID,
		TaskName:    m.TaskName,
		CampaignID:  m.CampaignID,
		CampaignName: m.CampaignName,
		UserID:      m.UserID,
		TenantID:    m.TenantID,
		AppID:       m.AppID,
		GrantID:     m.GrantID,
		ProgressData: string(m.ProgressData),
		TriggerEvent: m.TriggerEvent,
		CompletedAt: m.CompletedAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:  m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *taskCompletionLogRepo) toDataModel(b *biz.TaskCompletionLog) *model.TaskCompletionLog {
	if b == nil {
		return nil
	}
	return &model.TaskCompletionLog{
		CompletionID: b.CompletionID,
		TaskID:      b.TaskID,
		TaskName:    b.TaskName,
		CampaignID:  b.CampaignID,
		CampaignName: b.CampaignName,
		UserID:      b.UserID,
		TenantID:    b.TenantID,
		AppID:       b.AppID,
		GrantID:     b.GrantID,
		ProgressData: []byte(b.ProgressData),
		TriggerEvent: b.TriggerEvent,
		CompletedAt: b.CompletedAt,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
}

// Save 保存任务完成记录
func (r *taskCompletionLogRepo) Save(ctx context.Context, log *biz.TaskCompletionLog) (*biz.TaskCompletionLog, error) {
	m := r.toDataModel(log)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save task completion log: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// FindByID 根据ID查找任务完成记录
func (r *taskCompletionLogRepo) FindByID(ctx context.Context, id string) (*biz.TaskCompletionLog, error) {
	var m model.TaskCompletionLog
	if err := r.data.db.WithContext(ctx).Where("completion_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("failed to find task completion log by id: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// List 列出任务完成记录（分页）
func (r *taskCompletionLogRepo) List(ctx context.Context, tenantID, appID, taskID, campaignID string, userID int64, page, pageSize int) ([]*biz.TaskCompletionLog, int64, error) {
	var (
		models []model.TaskCompletionLog
		total   int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.TaskCompletionLog{})

	// 添加过滤条件
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count task completion logs: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("completed_at DESC").Find(&models).Error; err != nil {
		r.log.Errorf("failed to list task completion logs: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.TaskCompletionLog, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// CountByTaskAndUser 统计用户完成某个任务的次数
func (r *taskCompletionLogRepo) CountByTaskAndUser(ctx context.Context, taskID string, userID int64) (int64, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&model.TaskCompletionLog{}).
		Where("task_id = ? AND user_id = ?", taskID, userID).
		Count(&count).Error; err != nil {
		r.log.Errorf("failed to count task completions: %v", err)
		return 0, err
	}
	return count, nil
}

// CountByTask 统计任务总完成次数
func (r *taskCompletionLogRepo) CountByTask(ctx context.Context, taskID, campaignID string) (int64, error) {
	var count int64
	query := r.data.db.WithContext(ctx).Model(&model.TaskCompletionLog{}).
		Where("task_id = ?", taskID)
	
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}
	
	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("failed to count task completions: %v", err)
		return 0, err
	}
	return count, nil
}

// CountUniqueUsersByTask 统计任务的唯一用户数
func (r *taskCompletionLogRepo) CountUniqueUsersByTask(ctx context.Context, taskID, campaignID string) (int64, error) {
	var count int64
	query := r.data.db.WithContext(ctx).Model(&model.TaskCompletionLog{}).
		Where("task_id = ?", taskID)
	
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}
	
	if err := query.Distinct("user_id").Count(&count).Error; err != nil {
		r.log.Errorf("failed to count unique users: %v", err)
		return 0, err
	}
	return count, nil
}

