package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// redeemCodeRepo 实现 biz.RedeemCodeRepo 接口
type redeemCodeRepo struct {
	data *Data
	log  *log.Helper
}

// NewRedeemCodeRepo 创建 RedeemCode Repository
func NewRedeemCodeRepo(data *Data, logger log.Logger) biz.RedeemCodeRepo {
	return &redeemCodeRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *redeemCodeRepo) toBizModel(m *model.RedeemCode) *biz.RedeemCode {
	if m == nil {
		return nil
	}
	return &biz.RedeemCode{
		Code:         m.Code,
		TenantID:     m.TenantID,
		AppID:        m.AppID,
		GrantID:      m.GrantID,
		CampaignID:   m.CampaignID,
		CampaignName: m.CampaignName,
		RewardID:     m.RewardID,
		RewardName:   m.RewardName,
		BatchID:      m.BatchID,
		Status:       m.Status,
		OwnerUserID:  m.OwnerUserID,
		RedeemedBy:   m.RedeemedBy,
		RedeemedAt:   m.RedeemedAt,
		ExpireAt:     m.ExpireAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *redeemCodeRepo) toDataModel(b *biz.RedeemCode) *model.RedeemCode {
	if b == nil {
		return nil
	}
	return &model.RedeemCode{
		Code:         b.Code,
		TenantID:     b.TenantID,
		AppID:        b.AppID,
		GrantID:      b.GrantID,
		CampaignID:   b.CampaignID,
		CampaignName: b.CampaignName,
		RewardID:     b.RewardID,
		RewardName:   b.RewardName,
		BatchID:      b.BatchID,
		Status:       b.Status,
		OwnerUserID:  b.OwnerUserID,
		RedeemedBy:   b.RedeemedBy,
		RedeemedAt:   b.RedeemedAt,
		ExpireAt:     b.ExpireAt,
		CreatedAt:    b.CreatedAt,
		UpdatedAt:    b.UpdatedAt,
	}
}

// Save 保存兑换码（创建或更新）
func (r *redeemCodeRepo) Save(ctx context.Context, rc *biz.RedeemCode) (*biz.RedeemCode, error) {
	m := r.toDataModel(rc)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save redeem code: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// FindByCode 根据兑换码查找
func (r *redeemCodeRepo) FindByCode(ctx context.Context, code, tenantID string) (*biz.RedeemCode, error) {
	var m model.RedeemCode
	if err := r.data.db.WithContext(ctx).
		Where("code = ? AND tenant_id = ?", code, tenantID).
		First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("failed to find redeem code by code: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// List 列出兑换码（分页）
func (r *redeemCodeRepo) List(ctx context.Context, tenantID, appID, campaignID, batchID string, userID int64, status string, page, pageSize int) ([]*biz.RedeemCode, int64, error) {
	var (
		models []model.RedeemCode
		total  int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.RedeemCode{})

	// 添加过滤条件
	if tenantID != "" {
		query = query.Where("tenant_id = ?", tenantID)
	}
	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}
	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}
	if userID > 0 {
		query = query.Where("(owner_user_id = ? OR redeemed_by = ?)", userID, userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count redeem codes: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		r.log.Errorf("failed to list redeem codes: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.RedeemCode, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// UpdateStatus 更新状态
func (r *redeemCodeRepo) UpdateStatus(ctx context.Context, code, tenantID, status string) error {
	if err := r.data.db.WithContext(ctx).Model(&model.RedeemCode{}).
		Where("code = ? AND tenant_id = ?", code, tenantID).
		Update("status", status).Error; err != nil {
		r.log.Errorf("failed to update redeem code status: %v", err)
		return err
	}
	return nil
}

// Redeem 兑换码核销
func (r *redeemCodeRepo) Redeem(ctx context.Context, code, tenantID string, userID int64) error {
	now := r.data.db.NowFunc()
	if err := r.data.db.WithContext(ctx).Model(&model.RedeemCode{}).
		Where("code = ? AND tenant_id = ? AND status = ?", code, tenantID, model.RedeemCodeStatusActive).
		Updates(map[string]interface{}{
			"status":      model.RedeemCodeStatusRedeemed,
			"redeemed_by": userID,
			"redeemed_at": now,
		}).Error; err != nil {
		r.log.Errorf("failed to redeem code: %v", err)
		return err
	}
	return nil
}

// BatchCreate 批量创建兑换码（性能优化版本）
func (r *redeemCodeRepo) BatchCreate(ctx context.Context, codes []*biz.RedeemCode) error {
	if len(codes) == 0 {
		return nil
	}

	// 性能优化：
	// 1. 预分配切片容量，避免多次扩容
	models := make([]*model.RedeemCode, 0, len(codes))

	// 2. 批量转换，减少内存分配
	for _, c := range codes {
		models = append(models, r.toDataModel(c))
	}

	// 3. 使用事务批量插入，提高性能
	// 根据数据量动态调整批次大小
	batchSize := 500 // 增大批次大小以提高性能
	if len(models) < batchSize {
		batchSize = len(models)
	}

	// 4. 使用 CreateInBatches 进行批量插入
	// GORM 会自动处理事务和错误回滚
	if err := r.data.db.WithContext(ctx).CreateInBatches(models, batchSize).Error; err != nil {
		r.log.Errorf("failed to batch create redeem codes: %v", err)
		return err
	}

	r.log.Infof("batch created %d redeem codes in batches of %d", len(models), batchSize)
	return nil
}
