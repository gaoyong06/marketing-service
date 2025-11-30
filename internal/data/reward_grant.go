package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// rewardGrantRepo 实现 biz.RewardGrantRepo 接口
type rewardGrantRepo struct {
	data *Data
	log  *log.Helper
}

// NewRewardGrantRepo 创建 RewardGrant Repository
func NewRewardGrantRepo(data *Data, logger log.Logger) biz.RewardGrantRepo {
	return &rewardGrantRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *rewardGrantRepo) toBizModel(m *model.RewardGrant) *biz.RewardGrant {
	if m == nil {
		return nil
	}
	return &biz.RewardGrant{
		GrantID:         m.GrantID,
		RewardID:        m.RewardID,
		RewardName:      m.RewardName,
		RewardType:      m.RewardType,
		RewardVersion:   m.RewardVersion,
		ContentSnapshot: string(m.ContentSnapshot),
		GeneratorConfig: string(m.GeneratorConfig),
		CampaignID:      m.CampaignID,
		CampaignName:    m.CampaignName,
		TaskID:          m.TaskID,
		TaskName:        m.TaskName,
		TenantID:        m.TenantID,
		AppID:           m.AppID,
		UserID:          m.UserID,
		Status:          m.Status,
		ReservedAt:      m.ReservedAt,
		DistributedAt:   m.DistributedAt,
		UsedAt:          m.UsedAt,
		ExpireTime:      m.ExpireTime,
		ErrorMessage:    m.ErrorMessage,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *rewardGrantRepo) toDataModel(b *biz.RewardGrant) *model.RewardGrant {
	if b == nil {
		return nil
	}
	return &model.RewardGrant{
		GrantID:         b.GrantID,
		RewardID:        b.RewardID,
		RewardName:      b.RewardName,
		RewardType:      b.RewardType,
		RewardVersion:   b.RewardVersion,
		ContentSnapshot: []byte(b.ContentSnapshot),
		GeneratorConfig: []byte(b.GeneratorConfig),
		CampaignID:      b.CampaignID,
		CampaignName:    b.CampaignName,
		TaskID:          b.TaskID,
		TaskName:        b.TaskName,
		TenantID:        b.TenantID,
		AppID:           b.AppID,
		UserID:          b.UserID,
		Status:          b.Status,
		ReservedAt:      b.ReservedAt,
		DistributedAt:   b.DistributedAt,
		UsedAt:          b.UsedAt,
		ExpireTime:      b.ExpireTime,
		ErrorMessage:    b.ErrorMessage,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

// Save 保存奖励发放记录
func (r *rewardGrantRepo) Save(ctx context.Context, grant *biz.RewardGrant) (*biz.RewardGrant, error) {
	m := r.toDataModel(grant)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save reward grant: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// Update 更新奖励发放记录
func (r *rewardGrantRepo) Update(ctx context.Context, grant *biz.RewardGrant) (*biz.RewardGrant, error) {
	m := r.toDataModel(grant)
	if err := r.data.db.WithContext(ctx).Model(&model.RewardGrant{}).
		Where("grant_id = ?", m.GrantID).Updates(m).Error; err != nil {
		r.log.Errorf("failed to update reward grant: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据
	return r.FindByID(ctx, m.GrantID)
}

// FindByID 根据ID查找奖励发放记录
func (r *rewardGrantRepo) FindByID(ctx context.Context, id string) (*biz.RewardGrant, error) {
	var m model.RewardGrant
	if err := r.data.db.WithContext(ctx).Where("grant_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("failed to find reward grant by id: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// List 列出奖励发放记录（分页，性能优化版本）
func (r *rewardGrantRepo) List(ctx context.Context, tenantID, appID string, userID int64, status string, page, pageSize int) ([]*biz.RewardGrant, int64, error) {
	var (
		models []model.RewardGrant
		total  int64
	)

	// 性能优化：使用复合索引 idx_tenant_app_user_status
	query := r.data.db.WithContext(ctx).Model(&model.RewardGrant{})

	// 添加过滤条件（利用复合索引，按索引顺序添加条件）
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 性能优化：使用 Select 只查询需要的字段（减少 JSON 字段的传输）
	query = query.Select("grant_id, reward_id, reward_name, reward_type, reward_version, content_snapshot, campaign_id, campaign_name, task_id, task_name, tenant_id, app_id, user_id, status, distributed_at, used_at, expire_time, created_at, updated_at")

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count reward grants: %v", err)
		return nil, 0, err
	}

	// 分页查询（性能优化：深度分页警告）
	offset := (page - 1) * pageSize
	if offset > 10000 {
		r.log.Warnf("deep pagination detected: page=%d, consider using cursor-based pagination", page)
	}

	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC, grant_id DESC"). // 添加 grant_id 确保排序稳定
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list reward grants: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型（预分配容量）
	result := make([]*biz.RewardGrant, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// UpdateStatus 更新状态
func (r *rewardGrantRepo) UpdateStatus(ctx context.Context, grantID, status string) error {
	if err := r.data.db.WithContext(ctx).Model(&model.RewardGrant{}).
		Where("grant_id = ?", grantID).
		Update("status", status).Error; err != nil {
		r.log.Errorf("failed to update reward grant status: %v", err)
		return err
	}
	return nil
}

// CountByStatus 按状态统计数量
func (r *rewardGrantRepo) CountByStatus(ctx context.Context, rewardID, status string) (int64, error) {
	var count int64
	query := r.data.db.WithContext(ctx).Model(&model.RewardGrant{}).
		Where("reward_id = ?", rewardID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&count).Error; err != nil {
		r.log.Errorf("failed to count reward grants by status: %v", err)
		return 0, err
	}
	return count, nil
}

// BatchSave 批量保存奖励发放记录（性能优化版本）
func (r *rewardGrantRepo) BatchSave(ctx context.Context, grants []*biz.RewardGrant) error {
	if len(grants) == 0 {
		return nil
	}

	// 性能优化：
	// 1. 预分配切片容量
	models := make([]*model.RewardGrant, 0, len(grants))

	// 2. 批量转换
	for _, g := range grants {
		models = append(models, r.toDataModel(g))
	}

	// 3. 根据数据量动态调整批次大小
	batchSize := 500 // 增大批次大小以提高性能
	if len(models) < batchSize {
		batchSize = len(models)
	}

	// 4. 使用 CreateInBatches 进行批量插入
	if err := r.data.db.WithContext(ctx).CreateInBatches(models, batchSize).Error; err != nil {
		r.log.Errorf("failed to batch save reward grants: %v", err)
		return err
	}

	r.log.Infof("batch saved %d reward grants in batches of %d", len(models), batchSize)
	return nil
}
