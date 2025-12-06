package data

import (
	"context"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"
	"time"

	"github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// couponRepo 实现 biz.CouponRepo 接口
type couponRepo struct {
	data *Data
	log  *log.Helper
}

// NewCouponRepo 创建 Coupon Repository
func NewCouponRepo(data *Data, logger log.Logger) biz.CouponRepo {
	return &couponRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

// toBizModel 将数据模型转换为业务模型
func (r *couponRepo) toBizModel(m *model.Coupon) *biz.Coupon {
	if m == nil {
		return nil
	}
	return &biz.Coupon{
		Code:          m.Code,
		AppID:         m.AppID,
		DiscountType:  m.DiscountType,
		DiscountValue: m.DiscountValue,
		ValidFrom:     m.ValidFrom,
		ValidUntil:    m.ValidUntil,
		MaxUses:       m.MaxUses,
		UsedCount:     m.UsedCount,
		MinAmount:     m.MinAmount,
		Status:        m.Status,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// toDataModel 将业务模型转换为数据模型
func (r *couponRepo) toDataModel(b *biz.Coupon) *model.Coupon {
	if b == nil {
		return nil
	}
	return &model.Coupon{
		Code:          b.Code,
		AppID:         b.AppID,
		DiscountType:  b.DiscountType,
		DiscountValue: b.DiscountValue,
		ValidFrom:     b.ValidFrom,
		ValidUntil:    b.ValidUntil,
		MaxUses:       b.MaxUses,
		UsedCount:     b.UsedCount,
		MinAmount:     b.MinAmount,
		Status:        b.Status,
		CreatedAt:     b.CreatedAt,
		UpdatedAt:     b.UpdatedAt,
	}
}

// toBizUsageModel 将使用记录数据模型转换为业务模型
func (r *couponRepo) toBizUsageModel(m *model.CouponUsage) *biz.CouponUsage {
	if m == nil {
		return nil
	}
	return &biz.CouponUsage{
		ID:             m.ID,
		CouponCode:     m.CouponCode,
		UserID:         m.UserID,
		OrderID:        m.OrderID,
		PaymentID:      m.PaymentID,
		OriginalAmount: m.OriginalAmount,
		DiscountAmount: m.DiscountAmount,
		FinalAmount:    m.FinalAmount,
		UsedAt:         m.UsedAt,
		CreatedAt:      m.CreatedAt,
	}
}

// toDataUsageModel 将使用记录业务模型转换为数据模型
func (r *couponRepo) toDataUsageModel(b *biz.CouponUsage) *model.CouponUsage {
	if b == nil {
		return nil
	}
	return &model.CouponUsage{
		ID:             b.ID,
		CouponCode:     b.CouponCode,
		UserID:         b.UserID,
		OrderID:        b.OrderID,
		PaymentID:      b.PaymentID,
		OriginalAmount: b.OriginalAmount,
		DiscountAmount: b.DiscountAmount,
		FinalAmount:    b.FinalAmount,
		UsedAt:         b.UsedAt,
		CreatedAt:      b.CreatedAt,
	}
}

// Save 保存优惠券（创建或更新）
func (r *couponRepo) Save(ctx context.Context, coupon *biz.Coupon) (*biz.Coupon, error) {
	m := r.toDataModel(coupon)
	if err := r.data.db.WithContext(ctx).Save(m).Error; err != nil {
		r.log.Errorf("failed to save coupon: %v", err)
		return nil, err
	}
	return r.toBizModel(m), nil
}

// Update 更新优惠券
func (r *couponRepo) Update(ctx context.Context, coupon *biz.Coupon) (*biz.Coupon, error) {
	m := r.toDataModel(coupon)
	updateFields := map[string]interface{}{
		"discount_type":  m.DiscountType,
		"discount_value": m.DiscountValue,
		"valid_from":     m.ValidFrom,
		"valid_until":    m.ValidUntil,
		"max_uses":       m.MaxUses,
		"min_amount":     m.MinAmount,
		"status":         m.Status,
		"updated_at":     m.UpdatedAt,
	}
	if err := r.data.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("code = ?", m.Code).Updates(updateFields).Error; err != nil {
		r.log.Errorf("failed to update coupon: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据
	return r.FindByCode(ctx, m.Code)
}

// FindByCode 根据优惠码查找优惠券
func (r *couponRepo) FindByCode(ctx context.Context, code string) (*biz.Coupon, error) {
	var m model.Coupon
	if err := r.data.db.WithContext(ctx).Where("code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NewBizError(errors.ErrCodeNotFound, "zh-CN")
		}
		r.log.Errorf("failed to find coupon by code: %v", err)
		return nil, err
	}
	return r.toBizModel(&m), nil
}

// List 列出优惠券（分页）
func (r *couponRepo) List(ctx context.Context, appID, status string, page, pageSize int) ([]*biz.Coupon, int64, error) {
	var (
		models []model.Coupon
		total  int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.Coupon{})

	if appID != "" {
		query = query.Where("app_id = ?", appID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count coupons: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC, code DESC").
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list coupons: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.Coupon, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// Delete 删除优惠券
func (r *couponRepo) Delete(ctx context.Context, code string) error {
	if err := r.data.db.WithContext(ctx).Where("code = ?", code).
		Delete(&model.Coupon{}).Error; err != nil {
		r.log.Errorf("failed to delete coupon: %v", err)
		return err
	}
	return nil
}

// IncrementUsedCount 原子性增加使用次数
func (r *couponRepo) IncrementUsedCount(ctx context.Context, code string) error {
	// 使用数据库的原子操作，同时检查是否超过最大使用次数
	// 注意：max_uses = 0 表示无限制，所以条件为 (max_uses = 0 OR used_count < max_uses)
	result := r.data.db.WithContext(ctx).Model(&model.Coupon{}).
		Where("code = ? AND (max_uses = 0 OR used_count < max_uses)", code).
		Update("used_count", gorm.Expr("used_count + 1"))

	if result.Error != nil {
		r.log.Errorf("failed to increment used count: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.NewBizError(errors.ErrCodeBusinessRuleViolation, "zh-CN")
	}

	return nil
}

// CreateUsage 创建使用记录
func (r *couponRepo) CreateUsage(ctx context.Context, usage *biz.CouponUsage) error {
	m := r.toDataUsageModel(usage)
	if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
		r.log.Errorf("failed to create coupon usage: %v", err)
		return err
	}
	return nil
}

// UseCoupon 使用优惠券（事务操作：原子性增加使用次数 + 创建使用记录）
func (r *couponRepo) UseCoupon(ctx context.Context, code string, userID uint64, orderID, paymentID string, originalAmount, discountAmount, finalAmount int64) error {
	// 使用事务确保原子性
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 原子性增加使用次数
		result := tx.Model(&model.Coupon{}).
			Where("code = ? AND (max_uses = 0 OR used_count < max_uses)", code).
			Update("used_count", gorm.Expr("used_count + 1"))

		if result.Error != nil {
			r.log.Errorf("failed to increment used count: %v", result.Error)
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.NewBizError(errors.ErrCodeBusinessRuleViolation, "zh-CN")
		}

		// 2. 创建使用记录
		usage := &model.CouponUsage{
			ID:             biz.GenerateShortID(),
			CouponCode:     code,
			UserID:         userID,
			OrderID:        orderID,
			PaymentID:      paymentID,
			OriginalAmount: originalAmount,
			DiscountAmount: discountAmount,
			FinalAmount:    finalAmount,
			UsedAt:         time.Now().Unix(),
			CreatedAt:      time.Now().Unix(),
		}

		if err := tx.Create(usage).Error; err != nil {
			r.log.Errorf("failed to create coupon usage: %v", err)
			return err
		}

		return nil
	})
}

// ListUsages 列出使用记录（分页）
func (r *couponRepo) ListUsages(ctx context.Context, couponCode string, page, pageSize int) ([]*biz.CouponUsage, int64, error) {
	var (
		models []model.CouponUsage
		total  int64
	)

	query := r.data.db.WithContext(ctx).Model(&model.CouponUsage{}).
		Where("coupon_code = ?", couponCode)

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		r.log.Errorf("failed to count coupon usages: %v", err)
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("used_at DESC, id DESC").
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list coupon usages: %v", err)
		return nil, 0, err
	}

	// 转换为业务模型
	result := make([]*biz.CouponUsage, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizUsageModel(&m))
	}

	return result, total, nil
}

// GetStats 获取优惠券统计
func (r *couponRepo) GetStats(ctx context.Context, code string) (*biz.CouponStats, error) {
	var stats biz.CouponStats
	stats.Code = code

	// 统计使用次数和订单数
	var countResult struct {
		TotalUses   int32
		TotalOrders int32
	}
	if err := r.data.db.WithContext(ctx).Model(&model.CouponUsage{}).
		Select("COUNT(*) as total_uses, COUNT(DISTINCT order_id) as total_orders").
		Where("coupon_code = ?", code).
		Scan(&countResult).Error; err != nil {
		r.log.Errorf("failed to count coupon stats: %v", err)
		return nil, err
	}
	stats.TotalUses = countResult.TotalUses
	stats.TotalOrders = countResult.TotalOrders

	// 统计收入和折扣金额
	var amountResult struct {
		TotalRevenue  int64
		TotalDiscount int64
	}
	if err := r.data.db.WithContext(ctx).Model(&model.CouponUsage{}).
		Select("SUM(final_amount) as total_revenue, SUM(discount_amount) as total_discount").
		Where("coupon_code = ?", code).
		Scan(&amountResult).Error; err != nil {
		r.log.Errorf("failed to sum coupon amounts: %v", err)
		return nil, err
	}
	stats.TotalRevenue = amountResult.TotalRevenue
	stats.TotalDiscount = amountResult.TotalDiscount

	// 计算转化率（如果有优惠券信息）
	var coupon model.Coupon
	if err := r.data.db.WithContext(ctx).Where("code = ?", code).First(&coupon).Error; err == nil {
		if coupon.MaxUses > 0 {
			stats.ConversionRate = float32(stats.TotalUses) / float32(coupon.MaxUses) * 100
		}
	}

	return &stats, nil
}

// GetSummaryStats 获取汇总统计
func (r *couponRepo) GetSummaryStats(ctx context.Context, appID string) (*biz.SummaryStats, error) {
	var stats biz.SummaryStats

	// 构建查询条件
	couponQuery := r.data.db.WithContext(ctx).Model(&model.Coupon{})
	usageQuery := r.data.db.WithContext(ctx).Model(&model.CouponUsage{})

	if appID != "" {
		couponQuery = couponQuery.Where("app_id = ?", appID)
		usageQuery = usageQuery.Where("coupon_code IN (SELECT code FROM coupon WHERE app_id = ?)", appID)
	}

	// 统计优惠券总数和激活数
	var couponCounts struct {
		Total  int64
		Active int64
	}
	if err := couponQuery.Select("COUNT(*) as total, SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) as active").
		Scan(&couponCounts).Error; err != nil {
		r.log.Errorf("failed to count coupons: %v", err)
		return nil, err
	}
	stats.TotalCoupons = int32(couponCounts.Total)
	stats.ActiveCoupons = int32(couponCounts.Active)

	// 统计总使用次数和订单数（需要重新构建查询）
	usageCountsQuery := r.data.db.WithContext(ctx).Model(&model.CouponUsage{})
	if appID != "" {
		usageCountsQuery = usageCountsQuery.Where("coupon_code IN (SELECT code FROM coupon WHERE app_id = ?)", appID)
	}
	var usageCounts struct {
		TotalUses   int32
		TotalOrders int32
	}
	if err := usageCountsQuery.Select("COUNT(*) as total_uses, COUNT(DISTINCT order_id) as total_orders").
		Scan(&usageCounts).Error; err != nil {
		r.log.Errorf("failed to count usages: %v", err)
		return nil, err
	}
	stats.TotalUses = usageCounts.TotalUses
	stats.TotalOrders = usageCounts.TotalOrders

	// 统计总收入和总折扣（需要重新构建查询）
	amountsQuery := r.data.db.WithContext(ctx).Model(&model.CouponUsage{})
	if appID != "" {
		amountsQuery = amountsQuery.Where("coupon_code IN (SELECT code FROM coupon WHERE app_id = ?)", appID)
	}
	var amounts struct {
		TotalRevenue  int64
		TotalDiscount int64
	}
	if err := amountsQuery.Select("COALESCE(SUM(final_amount), 0) as total_revenue, COALESCE(SUM(discount_amount), 0) as total_discount").
		Scan(&amounts).Error; err != nil {
		r.log.Errorf("failed to sum amounts: %v", err)
		return nil, err
	}
	stats.TotalRevenue = amounts.TotalRevenue
	stats.TotalDiscount = amounts.TotalDiscount

	// 批量获取所有优惠券的统计信息（优化：避免 N+1 查询）
	// 使用 JOIN 查询一次性获取所有优惠券的统计
	type CouponStatsResult struct {
		Code           string
		MaxUses        int32
		TotalUses      int32
		TotalOrders    int32
		TotalRevenue   int64
		TotalDiscount  int64
		ConversionRate float32
	}

	var statsResults []CouponStatsResult
	statsQuery := r.data.db.WithContext(ctx).
		Table("coupon c").
		Select(`
			c.code,
			c.max_uses,
			COALESCE(COUNT(cu.id), 0) as total_uses,
			COALESCE(COUNT(DISTINCT cu.order_id), 0) as total_orders,
			COALESCE(SUM(cu.final_amount), 0) as total_revenue,
			COALESCE(SUM(cu.discount_amount), 0) as total_discount,
			CASE 
				WHEN c.max_uses > 0 THEN (COALESCE(COUNT(cu.id), 0) * 100.0 / c.max_uses)
				ELSE 0
			END as conversion_rate
		`).
		Joins("LEFT JOIN coupon_usage cu ON c.code = cu.coupon_code")

	if appID != "" {
		statsQuery = statsQuery.Where("c.app_id = ?", appID)
	}

	statsQuery = statsQuery.Group("c.code, c.max_uses").
		Order("total_uses DESC").
		Limit(10) // 只取前10个

	if err := statsQuery.Scan(&statsResults).Error; err != nil {
		r.log.Errorf("failed to get coupon stats: %v", err)
		return nil, err
	}

	// 转换为业务模型
	topCoupons := make([]*biz.CouponStats, 0, len(statsResults))
	var totalConversionRate float32
	var validCouponsCount int32

	for _, sr := range statsResults {
		couponStats := &biz.CouponStats{
			Code:           sr.Code,
			TotalUses:      sr.TotalUses,
			TotalOrders:    sr.TotalOrders,
			TotalRevenue:   sr.TotalRevenue,
			TotalDiscount:  sr.TotalDiscount,
			ConversionRate: sr.ConversionRate,
		}
		topCoupons = append(topCoupons, couponStats)

		if sr.MaxUses > 0 {
			totalConversionRate += sr.ConversionRate
			validCouponsCount++
		}
	}

	// 计算平均转化率
	if validCouponsCount > 0 {
		stats.AverageConversionRate = totalConversionRate / float32(validCouponsCount)
	}

	stats.TopCoupons = topCoupons

	return &stats, nil
}
