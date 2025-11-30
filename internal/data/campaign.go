package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// campaignRepo 实现 biz.CampaignRepo 接口
type campaignRepo struct {
	data  *Data
	cache *CacheService
	log   *log.Helper
}

// NewCampaignRepo 创建 Campaign Repository
func NewCampaignRepo(data *Data, cache *CacheService, logger log.Logger) biz.CampaignRepo {
	return &campaignRepo{
		data:  data,
		cache: cache,
		log:   log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *campaignRepo) toBizModel(m *model.Campaign) *biz.Campaign {
	if m == nil {
		return nil
	}
	return &biz.Campaign{
		ID:              m.CampaignID,
		TenantID:        m.TenantID,
		AppID:           m.AppID,
		Name:            m.CampaignName,
		Type:            m.CampaignType,
		StartTime:       m.StartTime,
		EndTime:         m.EndTime,
		AudienceConfig:  string(m.AudienceConfig),
		ValidatorConfig: string(m.ValidatorConfig),
		Status:          m.Status,
		Description:     m.Description,
		CreatedBy:       m.CreatedBy,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *campaignRepo) toDataModel(b *biz.Campaign) *model.Campaign {
	if b == nil {
		return nil
	}
	return &model.Campaign{
		CampaignID:      b.ID,
		TenantID:        b.TenantID,
		AppID:           b.AppID,
		CampaignName:    b.Name,
		CampaignType:    b.Type,
		StartTime:       b.StartTime,
		EndTime:         b.EndTime,
		AudienceConfig:  []byte(b.AudienceConfig),
		ValidatorConfig: []byte(b.ValidatorConfig),
		Status:          b.Status,
		Description:     b.Description,
		CreatedBy:       b.CreatedBy,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

// Save 保存活动（创建或更新）
func (r *campaignRepo) Save(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	m := r.toDataModel(c)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save campaign: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// Update 更新活动
func (r *campaignRepo) Update(ctx context.Context, c *biz.Campaign) (*biz.Campaign, error) {
	m := r.toDataModel(c)
	if err := r.data.db.WithContext(ctx).Model(&model.Campaign{}).
		Where("campaign_id = ?", m.CampaignID).Updates(m).Error; err != nil {
		r.log.Errorf("failed to update campaign: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据（会自动更新缓存）
	result, err := r.FindByID(ctx, m.CampaignID)
	if err != nil {
		return nil, err
	}
	// 失效相关的任务列表缓存
	if r.cache != nil {
		_ = r.cache.InvalidateCampaignTasks(ctx, m.CampaignID)
	}
	return result, nil
}

// FindByID 根据ID查找活动（带缓存）
func (r *campaignRepo) FindByID(ctx context.Context, id string) (*biz.Campaign, error) {
	// 1. 先尝试从缓存获取
	if r.cache != nil {
		cached, err := r.cache.GetCampaign(ctx, id)
		if err == nil && cached != nil {
			r.log.Debugf("campaign cache hit: %s", id)
			return cached, nil
		}
		r.log.Debugf("campaign cache miss: %s", id)
	}

	// 2. 从数据库查询
	var m model.Campaign
	if err := r.data.db.WithContext(ctx).Where("campaign_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// 缓存穿透保护：对于不存在的记录，也缓存一个空值（短时间）
			if r.cache != nil {
				// 创建一个空的 Campaign 对象用于缓存穿透保护
				emptyCampaign := &biz.Campaign{ID: id}
				_ = r.cache.SetCampaign(ctx, emptyCampaign, 5*time.Minute) // 短时间缓存空值
			}
			return nil, nil
		}
		r.log.Errorf("failed to find campaign by id: %v", err)
		return nil, err
	}

	// 3. 转换为业务模型
	result := r.toBizModel(&m)

	// 4. 写入缓存
	if r.cache != nil && result != nil {
		if err := r.cache.SetCampaign(ctx, result, CampaignCacheTTL); err != nil {
			r.log.Warnf("failed to set campaign cache: %v", err)
		}
	}

	return result, nil
}

// List 列出活动（分页，性能优化版本）
func (r *campaignRepo) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*biz.Campaign, int64, error) {
	var (
		models []model.Campaign
		total  int64
	)

	// 性能优化：使用复合索引 idx_tenant_app_status
	query := r.data.db.WithContext(ctx).Model(&model.Campaign{})

	// 添加过滤条件（利用复合索引）
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	// 性能优化：使用 Select 只查询需要的字段，减少数据传输
	// 对于列表查询，可以只查询必要字段
	query = query.Select("campaign_id, tenant_id, app_id, campaign_name, campaign_type, start_time, end_time, status, created_at, updated_at")

	// 统计总数（优化：使用 COUNT(*) 而不是 COUNT(column)）
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count campaigns: %v", err)
		return nil, 0, err
	}

	// 性能优化：对于深度分页（page > 100），建议使用游标分页
	// 这里先使用 OFFSET/LIMIT，后续可以优化为游标分页
	offset := (page - 1) * pageSize
	if offset > 10000 {
		r.log.Warnf("deep pagination detected: page=%d, consider using cursor-based pagination", page)
	}

	// 分页查询（使用索引优化排序）
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC, campaign_id DESC"). // 添加 campaign_id 确保排序稳定
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list campaigns: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型（预分配容量）
	result := make([]*biz.Campaign, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// Delete 删除活动（软删除）
func (r *campaignRepo) Delete(ctx context.Context, id string) error {
	if err := r.data.db.WithContext(ctx).Where("campaign_id = ?", id).
		Delete(&model.Campaign{}).Error; err != nil {
		r.log.Errorf("failed to delete campaign: %v", err)
		return err
	}

	// 删除缓存
	if r.cache != nil {
		if err := r.cache.DeleteCampaign(ctx, id); err != nil {
			r.log.Warnf("failed to delete campaign cache: %v", err)
		}
		// 失效相关的任务列表缓存
		_ = r.cache.InvalidateCampaignTasks(ctx, id)
	}

	return nil
}
