package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// inventoryReservationRepo 实现 biz.InventoryReservationRepo 接口
type inventoryReservationRepo struct {
	data *Data
	log  *log.Helper
}

// NewInventoryReservationRepo 创建 InventoryReservation Repository
func NewInventoryReservationRepo(data *Data, logger log.Logger) biz.InventoryReservationRepo {
	return &inventoryReservationRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *inventoryReservationRepo) toBizModel(m *model.InventoryReservation) *biz.InventoryReservation {
	if m == nil {
		return nil
	}
	return &biz.InventoryReservation{
		ReservationID: m.ReservationID,
		ResourceID:    m.ResourceID,
		CampaignID:    m.CampaignID,
		UserID:        m.UserID,
		Quantity:      m.Quantity,
		Status:        m.Status,
		ExpireAt:      m.ExpireAt,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *inventoryReservationRepo) toDataModel(b *biz.InventoryReservation) *model.InventoryReservation {
	if b == nil {
		return nil
	}
	return &model.InventoryReservation{
		ReservationID: b.ReservationID,
		ResourceID:    b.ResourceID,
		CampaignID:    b.CampaignID,
		UserID:        b.UserID,
		Quantity:      b.Quantity,
		Status:        b.Status,
		ExpireAt:      b.ExpireAt,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

// Save 保存库存预占记录
func (r *inventoryReservationRepo) Save(ctx context.Context, ir *biz.InventoryReservation) (*biz.InventoryReservation, error) {
	m := r.toDataModel(ir)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save inventory reservation: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// FindByID 根据ID查找库存预占记录
func (r *inventoryReservationRepo) FindByID(ctx context.Context, id string) (*biz.InventoryReservation, error) {
	var m model.InventoryReservation
	if err := r.data.db.WithContext(ctx).Where("reservation_id = ?", id).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		r.log.Errorf("failed to find inventory reservation by id: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// UpdateStatus 更新状态
func (r *inventoryReservationRepo) UpdateStatus(ctx context.Context, reservationID, status string) error {
	if err := r.data.db.WithContext(ctx).Model(&model.InventoryReservation{}).
		Where("reservation_id = ?", reservationID).
		Update("status", status).Error; err != nil {
		r.log.Errorf("failed to update inventory reservation status: %v", err)
		return err
	}
	return nil
}

// CountPendingByResource 统计资源的待确认预占数量
func (r *inventoryReservationRepo) CountPendingByResource(ctx context.Context, resourceID string) (int, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&model.InventoryReservation{}).
		Where("resource_id = ? AND status = ?", resourceID, model.InventoryReservationStatusPending).
		Count(&count).Error; err != nil {
		r.log.Errorf("failed to count pending reservations: %v", err)
		return 0, err
	}
	return int(count), nil
}

// ListExpired 列出过期的预占记录
func (r *inventoryReservationRepo) ListExpired(ctx context.Context) ([]*biz.InventoryReservation, error) {
	var models []model.InventoryReservation
	now := r.data.db.NowFunc()
	if err := r.data.db.WithContext(ctx).
		Where("status = ? AND expire_at < ?", model.InventoryReservationStatusPending, now).
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list expired reservations: %v", err)
		return nil, err
	}

	result := make([]*biz.InventoryReservation, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}
	return result, nil
}

// CancelExpired 取消过期的预占
func (r *inventoryReservationRepo) CancelExpired(ctx context.Context) (int64, error) {
	now := r.data.db.NowFunc()
	result := r.data.db.WithContext(ctx).Model(&model.InventoryReservation{}).
		Where("status = ? AND expire_at < ?", model.InventoryReservationStatusPending, now).
		Update("status", model.InventoryReservationStatusExpired)
	if result.Error != nil {
		r.log.Errorf("failed to cancel expired reservations: %v", result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

// List 列出库存预占记录（分页）
func (r *inventoryReservationRepo) List(ctx context.Context, resourceID, campaignID string, userID int64, status string, page, pageSize int) ([]*biz.InventoryReservation, int64, error) {
	var (
		models []model.InventoryReservation
		total   int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.InventoryReservation{})

	// 添加过滤条件
	if resourceID != "" {
		query = query.Where("resource_id = ?", resourceID)
	}
	if campaignID != "" {
		query = query.Where("campaign_id = ?", campaignID)
	}
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count inventory reservations: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&models).Error; err != nil {
		r.log.Errorf("failed to list inventory reservations: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.InventoryReservation, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

