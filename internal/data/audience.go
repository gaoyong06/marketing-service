package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// audienceRepo 实现 biz.AudienceRepo 接口
type audienceRepo struct {
	data *Data
	log  *log.Helper
}

// NewAudienceRepo 创建 Audience Repository
func NewAudienceRepo(data *Data, logger log.Logger) biz.AudienceRepo {
	return &audienceRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *audienceRepo) toBizModel(m *model.Audience) *biz.Audience {
	if m == nil {
		return nil
	}
	return &biz.Audience{
		ID:           m.AudienceID,
		TenantID:     m.TenantID,
		AppID:        m.AppID,
		Name:         m.Name,
		AudienceType: m.AudienceType,
		RuleConfig:   string(m.RuleConfig),
		Status:       m.Status,
		Description:  m.Description,
		CreatedBy:    m.CreatedBy,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *audienceRepo) toDataModel(b *biz.Audience) *model.Audience {
	if b == nil {
		return nil
	}
	return &model.Audience{
		AudienceID:   b.ID,
		TenantID:     b.TenantID,
		AppID:        b.AppID,
		Name:         b.Name,
		AudienceType: b.AudienceType,
		RuleConfig:   []byte(b.RuleConfig),
		Status:       b.Status,
		Description:  b.Description,
		CreatedBy:    b.CreatedBy,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
}

// Save 保存受众（创建或更新）
func (r *audienceRepo) Save(ctx context.Context, a *biz.Audience) (*biz.Audience, error) {
	m := r.toDataModel(a)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save audience: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// Update 更新受众
func (r *audienceRepo) Update(ctx context.Context, a *biz.Audience) (*biz.Audience, error) {
	m := r.toDataModel(a)
	if err := r.data.db.WithContext(ctx).Model(&model.Audience{}).
		Where("audience_id = ?", m.AudienceID).Updates(m).Error; err != nil {
		r.log.Errorf("failed to update audience: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据
	return r.FindByID(ctx, m.AudienceID)
}

// FindByID 根据ID查找受众
func (r *audienceRepo) FindByID(ctx context.Context, id string) (*biz.Audience, error) {
	var m model.Audience
	if err := r.data.db.WithContext(ctx).Where("audience_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("failed to find audience by id: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// List 列出受众（分页）
func (r *audienceRepo) List(ctx context.Context, tenantID, appID string, page, pageSize int) ([]*biz.Audience, int64, error) {
	var (
		models []model.Audience
		total   int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.Audience{})

	// 添加过滤条件
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count audiences: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		r.log.Errorf("failed to list audiences: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.Audience, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// Delete 删除受众（软删除）
func (r *audienceRepo) Delete(ctx context.Context, id string) error {
	if err := r.data.db.WithContext(ctx).Where("audience_id = ?", id).
		Delete(&model.Audience{}).Error; err != nil {
		r.log.Errorf("failed to delete audience: %v", err)
		return err
	}
	return nil
}

