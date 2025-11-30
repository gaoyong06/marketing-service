package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// taskRepo 实现 biz.TaskRepo 接口
type taskRepo struct {
	data  *Data
	cache *CacheService
	log   *log.Helper
}

// NewTaskRepo 创建 Task Repository
func NewTaskRepo(data *Data, cache *CacheService, logger log.Logger) biz.TaskRepo {
	return &taskRepo{
		data:  data,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *taskRepo) toBizModel(m *model.Task) *biz.Task {
	if m == nil {
		return nil
	}
	return &biz.Task{
		ID:              m.TaskID,
		TenantID:        m.TenantID,
		AppID:           m.AppID,
		Name:            m.Name,
		TaskType:        m.TaskType,
		TriggerConfig:   string(m.TriggerConfig),
		ConditionConfig: string(m.ConditionConfig),
		RewardID:        m.RewardID,
		Status:          m.Status,
		StartTime:       m.StartTime,
		EndTime:         m.EndTime,
		MaxCount:        m.MaxCount,
		Description:     m.Description,
		CreatedBy:       m.CreatedBy,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *taskRepo) toDataModel(b *biz.Task) *model.Task {
	if b == nil {
		return nil
	}
	return &model.Task{
		TaskID:          b.ID,
		TenantID:        b.TenantID,
		AppID:           b.AppID,
		Name:            b.Name,
		TaskType:        b.TaskType,
		TriggerConfig:   []byte(b.TriggerConfig),
		ConditionConfig: []byte(b.ConditionConfig),
		RewardID:        b.RewardID,
		Status:          b.Status,
		StartTime:       b.StartTime,
		EndTime:         b.EndTime,
		MaxCount:        b.MaxCount,
		Description:     b.Description,
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

// Save 保存任务（创建或更新）
func (r *taskRepo) Save(ctx context.Context, t *biz.Task) (*biz.Task, error) {
	m := r.toDataModel(t)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save task: %v", err)
		return nil, err
	}
	result := r.toBizModel(m)

	// 更新缓存
	if r.cache != nil && result != nil {
		if err := r.cache.SetTask(ctx, result, TaskCacheTTL); err != nil {
			r.log.Warnf("failed to set task cache: %v", err)
		}
		// 如果任务关联了活动，失效活动的任务列表缓存
		if result.RewardID != "" {
			// 这里可以通过 Reward 找到关联的 Campaign，简化处理
		}
	}

	return result, nil
}

// Update 更新任务
func (r *taskRepo) Update(ctx context.Context, t *biz.Task) (*biz.Task, error) {
	m := r.toDataModel(t)
	if err := r.data.db.WithContext(ctx).Model(&model.Task{}).
		Where("task_id = ?", m.TaskID).Updates(m).Error; err != nil {
		r.log.Errorf("failed to update task: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据（会自动更新缓存）
	return r.FindByID(ctx, m.TaskID)
}

// FindByID 根据ID查找任务（带缓存）
func (r *taskRepo) FindByID(ctx context.Context, id string) (*biz.Task, error) {
	// 1. 先尝试从缓存获取
	if r.cache != nil {
		cached, err := r.cache.GetTask(ctx, id)
		if err == nil && cached != nil {
			r.log.Debugf("task cache hit: %s", id)
			return cached, nil
		}
		r.log.Debugf("task cache miss: %s", id)
	}

	// 2. 从数据库查询
	var m model.Task
	if err := r.data.db.WithContext(ctx).Where("task_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 缓存穿透保护：对于不存在的记录，也缓存一个空值（短时间）
			if r.cache != nil {
				emptyTask := &biz.Task{ID: id}
				_ = r.cache.SetTask(ctx, emptyTask, 5*time.Minute) // 短时间缓存空值
			}
			return nil, nil
		}
		r.log.Errorf("failed to find task by id: %v", err)
		return nil, err
	}

	// 3. 转换为业务模型
	result := r.toBizModel(&m)

	// 4. 写入缓存
	if r.cache != nil && result != nil {
		if err := r.cache.SetTask(ctx, result, TaskCacheTTL); err != nil {
			r.log.Warnf("failed to set task cache: %v", err)
		}
	}

	return result, nil
}

// List 列出任务（分页）
func (r *taskRepo) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*biz.Task, int64, error) {
	var (
		models []model.Task
		total  int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.Task{})

	// 添加过滤条件
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count tasks: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		r.log.Errorf("failed to list tasks: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.Task, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// ListByCampaign 根据活动ID列出任务
func (r *taskRepo) ListByCampaign(ctx context.Context, campaignID string) ([]*biz.Task, error) {
	var tasks []model.Task
	if err := r.data.db.WithContext(ctx).
		Table("task").
		Joins("INNER JOIN campaign_task ON task.task_id = campaign_task.task_id").
		Where("campaign_task.campaign_id = ?", campaignID).
		Find(&tasks).Error; err != nil {
		r.log.Errorf("failed to list tasks by campaign: %v", err)
		return nil, err
	}

	result := make([]*biz.Task, 0, len(tasks))
	for _, t := range tasks {
		result = append(result, r.toBizModel(&t))
	}
	return result, nil
}

// ListActive 列出活跃任务
func (r *taskRepo) ListActive(ctx context.Context, tenantID, appID string) ([]*biz.Task, error) {
	var models []model.Task
	now := r.data.db.NowFunc()
	if err := r.data.db.WithContext(ctx).
		Where("tenant_id = ? AND app_id = ? AND status = ? AND start_time <= ? AND end_time >= ?",
			tenantID, appID, "ACTIVE", now, now).
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list active tasks: %v", err)
		return nil, err
	}

	result := make([]*biz.Task, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}
	return result, nil
}

// Delete 删除任务（软删除）
// Delete 删除任务（软删除）
func (r *taskRepo) Delete(ctx context.Context, id string) error {
	if err := r.data.db.WithContext(ctx).Where("task_id = ?", id).
		Delete(&model.Task{}).Error; err != nil {
		r.log.Errorf("failed to delete task: %v", err)
		return err
	}

	// 删除缓存
	if r.cache != nil {
		if err := r.cache.DeleteTask(ctx, id); err != nil {
			r.log.Warnf("failed to delete task cache: %v", err)
		}
	}

	return nil
}
