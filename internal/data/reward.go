package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// rewardRepo 实现 biz.RewardRepo 接口
type rewardRepo struct {
	data  *Data
	cache *CacheService
	log   *log.Helper
}

// NewRewardRepo 创建 Reward Repository
func NewRewardRepo(data *Data, cache *CacheService, logger log.Logger) biz.RewardRepo {
	return &rewardRepo{
		data:  data,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *rewardRepo) toBizModel(m *model.Reward) *biz.Reward {
	if m == nil {
		return nil
	}
	return &biz.Reward{
		ID:                m.RewardID,
		TenantID:          m.TenantID,
		AppID:             m.AppID,
		RewardType:        m.RewardType,
		Name:              m.Name,
		ContentConfig:     string(m.ContentConfig),
		GeneratorConfig:   string(m.GeneratorConfig),
		DistributorConfig: string(m.DistributorConfig),
		ValidatorConfig:   string(m.ValidatorConfig),
		Version:           m.Version,
		ValidDays:         m.ValidDays,
		ExtraConfig:       string(m.ExtraConfig),
		Status:            m.Status,
		Description:       m.Description,
		CreatedBy:         m.CreatedBy,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *rewardRepo) toDataModel(b *biz.Reward) *model.Reward {
	if b == nil {
		return nil
	}
	return &model.Reward{
		RewardID:          b.ID,
		TenantID:          b.TenantID,
		AppID:             b.AppID,
		RewardType:        b.RewardType,
		Name:              b.Name,
		ContentConfig:     []byte(b.ContentConfig),
		GeneratorConfig:   []byte(b.GeneratorConfig),
		DistributorConfig: []byte(b.DistributorConfig),
		ValidatorConfig:   []byte(b.ValidatorConfig),
		Version:           b.Version,
		ValidDays:         b.ValidDays,
		ExtraConfig:       []byte(b.ExtraConfig),
		Status:            b.Status,
		Description:       b.Description,
		CreatedBy:         b.CreatedBy,
		CreatedAt:         b.CreatedAt,
		UpdatedAt:         b.UpdatedAt,
	}
}

// Save 保存奖励（创建或更新）
func (r *rewardRepo) Save(ctx context.Context, reward *biz.Reward) (*biz.Reward, error) {
	m := r.toDataModel(reward)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save reward: %v", err)
		return nil, err
	}
	result := r.toBizModel(m)

	// 更新缓存
	if r.cache != nil && result != nil {
		if err := r.cache.SetReward(ctx, result, RewardCacheTTL); err != nil {
			r.log.Warnf("failed to set reward cache: %v", err)
		}
		// 失效相关的奖励发放记录缓存
		_ = r.cache.InvalidateRewardGrants(ctx, result.ID)
	}

	return result, nil
}

// Update 更新奖励（版本号自动递增）
func (r *rewardRepo) Update(ctx context.Context, reward *biz.Reward) (*biz.Reward, error) {
	// 先查询当前版本
	var current model.Reward
	if err := r.data.db.WithContext(ctx).Where("reward_id = ?", reward.ID).First(&current).Error; err != nil {
		r.log.Errorf("failed to find reward before update: %v", err)
		return nil, err
	}

	m := r.toDataModel(reward)
	// 版本号递增
	m.Version = current.Version + 1

	if err := r.data.db.WithContext(ctx).Model(&model.Reward{}).
		Where("reward_id = ?", m.RewardID).Updates(m).Error; err != nil {
		r.log.Errorf("failed to update reward: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据（会自动更新缓存）
	result, err := r.FindByID(ctx, m.RewardID)
	if err != nil {
		return nil, err
	}
	// 失效相关的奖励发放记录缓存
	if r.cache != nil {
		_ = r.cache.InvalidateRewardGrants(ctx, m.RewardID)
	}
	return result, nil
}

// FindByID 根据ID查找奖励（带缓存）
func (r *rewardRepo) FindByID(ctx context.Context, id string) (*biz.Reward, error) {
	// 1. 先尝试从缓存获取
	if r.cache != nil {
		cached, err := r.cache.GetReward(ctx, id)
		if err == nil && cached != nil {
			r.log.Debugf("reward cache hit: %s", id)
			return cached, nil
		}
		r.log.Debugf("reward cache miss: %s", id)
	}

	// 2. 从数据库查询
	var m model.Reward
	if err := r.data.db.WithContext(ctx).Where("reward_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 缓存穿透保护：对于不存在的记录，也缓存一个空值（短时间）
			if r.cache != nil {
				emptyReward := &biz.Reward{ID: id}
				_ = r.cache.SetReward(ctx, emptyReward, 5*time.Minute) // 短时间缓存空值
			}
			return nil, nil
		}
		r.log.Errorf("failed to find reward by id: %v", err)
		return nil, err
	}

	// 3. 转换为业务模型
	result := r.toBizModel(&m)

	// 4. 写入缓存
	if r.cache != nil && result != nil {
		if err := r.cache.SetReward(ctx, result, RewardCacheTTL); err != nil {
			r.log.Warnf("failed to set reward cache: %v", err)
		}
	}

	return result, nil
}

// List 列出奖励（分页，性能优化版本）
func (r *rewardRepo) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*biz.Reward, int64, error) {
	var (
		models []model.Reward
		total  int64
	)

	// 性能优化：使用复合索引 idx_tenant_app_status
	query := r.data.db.WithContext(ctx).Model(&model.Reward{})

	// 添加过滤条件（利用复合索引）
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	// 性能优化：使用 Select 只查询需要的字段
	query = query.Select("reward_id, tenant_id, app_id, reward_type, name, status, valid_days, created_at, updated_at")

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count rewards: %v", err)
		return nil, 0, err
	}

	// 分页查询（性能优化：深度分页警告）
	offset := (page - 1) * pageSize
	if offset > 10000 {
		r.log.Warnf("deep pagination detected: page=%d, consider using cursor-based pagination", page)
	}

	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC, reward_id DESC"). // 添加 reward_id 确保排序稳定
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list rewards: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型（预分配容量）
	result := make([]*biz.Reward, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// Delete 删除奖励（软删除）
func (r *rewardRepo) Delete(ctx context.Context, id string) error {
	if err := r.data.db.WithContext(ctx).Where("reward_id = ?", id).
		Delete(&model.Reward{}).Error; err != nil {
		r.log.Errorf("failed to delete reward: %v", err)
		return err
	}

	// 删除缓存
	if r.cache != nil {
		if err := r.cache.DeleteReward(ctx, id); err != nil {
			r.log.Warnf("failed to delete reward cache: %v", err)
		}
		// 失效相关的奖励发放记录缓存
		_ = r.cache.InvalidateRewardGrants(ctx, id)
	}

	return nil
}
