package data

import (
	"context"
	"errors"
	"marketing-service/internal/biz"
	"marketing-service/internal/data/model"
	"strings"
	"time"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-sql-driver/mysql"
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
		CouponID:      m.CouponID,
		CouponCode:    m.CouponCode,
		AppID:         m.AppID,
		DiscountType:  m.DiscountType,
		DiscountValue: m.DiscountValue,
		Currency:      m.Currency,
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
	// 如果货币单位为空，设置默认值为 CNY
	currency := b.Currency
	if currency == "" {
		currency = "CNY"
	}
	return &model.Coupon{
		CouponID:      b.CouponID,
		CouponCode:    b.CouponCode,
		AppID:         b.AppID,
		DiscountType:  b.DiscountType,
		DiscountValue: b.DiscountValue,
		Currency:      currency,
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
		CouponUsageID:  m.CouponUsageID,
		CouponCode:     m.CouponCode,
		AppID:          m.AppID,
		UserID:         m.UserID,
		PaymentOrderID: m.PaymentOrderID,
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
		CouponUsageID:  b.CouponUsageID,
		CouponCode:     b.CouponCode,
		AppID:          b.AppID,
		UserID:         b.UserID,
		PaymentOrderID: b.PaymentOrderID,
		PaymentID:      b.PaymentID,
		OriginalAmount: b.OriginalAmount,
		DiscountAmount: b.DiscountAmount,
		FinalAmount:    b.FinalAmount,
		UsedAt:         b.UsedAt,
		CreatedAt:      b.CreatedAt,
	}
}

// isDuplicateEntryError 检查是否是 MySQL 唯一约束冲突错误
func isDuplicateEntryError(err error) bool {
	if err == nil {
		return false
	}
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1062 // Duplicate entry
	}
	// 也检查错误字符串
	errStr := err.Error()
	return strings.Contains(errStr, "Duplicate entry") ||
		strings.Contains(errStr, "1062") ||
		strings.Contains(errStr, "UNIQUE constraint")
}

// Save 保存优惠券（创建或更新）
func (r *couponRepo) Save(ctx context.Context, coupon *biz.Coupon) (*biz.Coupon, error) {
	m := r.toDataModel(coupon)
	// 使用 Create 而不是 Save，因为 Save 会在主键存在时更新，不会触发重复键错误
	// 对于创建操作，应该使用 Create，这样在优惠码已存在时会返回重复键错误
	if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
		r.log.Errorf("failed to save coupon: %v, coupon data: coupon_code=%s, app_id=%s, discount_type=%s, discount_value=%d, valid_from=%d, valid_until=%d, max_uses=%d, used_count=%d, min_amount=%d, status=%s, created_at=%d, updated_at=%d",
			err, m.CouponCode, m.AppID, m.DiscountType, m.DiscountValue, m.ValidFrom, m.ValidUntil, m.MaxUses, m.UsedCount, m.MinAmount, m.Status, m.CreatedAt, m.UpdatedAt)
		// 检查是否是重复键错误（优惠码已存在）
		if isDuplicateEntryError(err) {
			// 检查是否存在已删除的相同 code 的优惠券（软删除）
			var deletedCoupon model.Coupon
			if err := r.data.db.WithContext(ctx).Unscoped().Where("coupon_code = ?", m.CouponCode).First(&deletedCoupon).Error; err == nil {
				// 如果存在已删除的记录，永久删除它，然后重新创建
				r.log.Infof("Found soft-deleted coupon with code %s, permanently deleting it before creating new one", m.CouponCode)
				if err := r.data.db.WithContext(ctx).Unscoped().Where("coupon_code = ?", m.CouponCode).Delete(&model.Coupon{}).Error; err != nil {
					r.log.Errorf("failed to permanently delete soft-deleted coupon: %v", err)
					return nil, pkgErrors.WrapErrorWithLang(ctx, err, pkgErrors.ErrCodeInternalError)
				}
				// 重新创建
				if err := r.data.db.WithContext(ctx).Create(m).Error; err != nil {
					r.log.Errorf("failed to create coupon after deleting soft-deleted one: %v", err)
					return nil, pkgErrors.WrapErrorWithLang(ctx, err, pkgErrors.ErrCodeInternalError)
				}
				return r.toBizModel(m), nil
			}
			// 如果不存在已删除的记录，说明是真正的重复（未删除的记录）
			return nil, pkgErrors.NewBizError(pkgErrors.ErrCodeAlreadyExists, "zh-CN")
		}
		return nil, pkgErrors.WrapErrorWithLang(ctx, err, pkgErrors.ErrCodeInternalError)
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
		Where("coupon_code = ?", m.CouponCode).Updates(updateFields).Error; err != nil {
		r.log.Errorf("failed to update coupon: %v", err)
		return nil, err
	}
	// 重新查询以获取最新数据
	return r.FindByCode(ctx, m.CouponCode)
}

// FindByCode 根据优惠码查找优惠券
func (r *couponRepo) FindByCode(ctx context.Context, code string) (*biz.Coupon, error) {
	var m model.Coupon
	if err := r.data.db.WithContext(ctx).Where("coupon_code = ?", code).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, pkgErrors.NewBizError(pkgErrors.ErrCodeNotFound, "zh-CN")
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
		return nil, 0, pkgErrors.WrapErrorWithLang(ctx, err, pkgErrors.ErrCodeInternalError)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).
		Order("created_at DESC, coupon_code DESC").
		Find(&models).Error; err != nil {
		r.log.Errorf("failed to list coupons: %v", err)
		return nil, 0, pkgErrors.WrapErrorWithLang(ctx, err, pkgErrors.ErrCodeInternalError)
	}

	// 转换为业务模型
	result := make([]*biz.Coupon, 0, len(models))
	for _, m := range models {
		result = append(result, r.toBizModel(&m))
	}

	return result, total, nil
}

// Delete 删除优惠券（软删除）
func (r *couponRepo) Delete(ctx context.Context, code string) error {
	// GORM 的软删除：使用 Delete 方法会自动设置 deleted_at 字段
	// 查询时会自动过滤 deleted_at IS NULL 的记录
	if err := r.data.db.WithContext(ctx).Where("coupon_code = ?", code).
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
		Where("coupon_code = ? AND (max_uses = 0 OR used_count < max_uses)", code).
		Update("used_count", gorm.Expr("used_count + 1"))

	if result.Error != nil {
		r.log.Errorf("failed to increment used count: %v", result.Error)
		return result.Error
	}

	if result.RowsAffected == 0 {
		return pkgErrors.NewBizError(pkgErrors.ErrCodeBusinessRuleViolation, "zh-CN")
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
func (r *couponRepo) UseCoupon(ctx context.Context, code, appID string, userID string, paymentOrderID, paymentID string, originalAmount, discountAmount, finalAmount int64) error {
	// 使用事务确保原子性
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 原子性增加使用次数
		result := tx.Model(&model.Coupon{}).
			Where("coupon_code = ? AND (max_uses = 0 OR used_count < max_uses)", code).
			Update("used_count", gorm.Expr("used_count + 1"))

		if result.Error != nil {
			r.log.Errorf("failed to increment used count: %v", result.Error)
			return result.Error
		}

		if result.RowsAffected == 0 {
			return pkgErrors.NewBizError(pkgErrors.ErrCodeBusinessRuleViolation, "zh-CN")
		}

		// 2. 创建使用记录
		now := time.Now()
		usage := &model.CouponUsage{
			CouponUsageID:  biz.GenerateShortID(),
			CouponCode:     code,
			AppID:          appID,
			UserID:         userID,
			PaymentOrderID: paymentOrderID,
			PaymentID:      paymentID,
			OriginalAmount: originalAmount,
			DiscountAmount: discountAmount,
			FinalAmount:    finalAmount,
			UsedAt:         now,
			CreatedAt:      now,
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
	stats.CouponCode = code

	// 统计使用次数和订单数
	var countResult struct {
		TotalUses   int32
		TotalOrders int32
	}
	if err := r.data.db.WithContext(ctx).Model(&model.CouponUsage{}).
		Select("COUNT(*) as total_uses, COUNT(DISTINCT payment_order_id) as total_orders").
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
	if err := r.data.db.WithContext(ctx).Where("coupon_code = ?", code).First(&coupon).Error; err == nil {
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
		usageQuery = usageQuery.Where("app_id = ?", appID)
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

	// 统计总使用次数和订单数（使用 app_id 字段直接查询，避免 JOIN）
	usageCountsQuery := r.data.db.WithContext(ctx).Model(&model.CouponUsage{})
	if appID != "" {
		usageCountsQuery = usageCountsQuery.Where("app_id = ?", appID)
	}
	var usageCounts struct {
		TotalUses   int32
		TotalOrders int32
	}
	if err := usageCountsQuery.Select("COUNT(*) as total_uses, COUNT(DISTINCT payment_order_id) as total_orders").
		Scan(&usageCounts).Error; err != nil {
		r.log.Errorf("failed to count usages: %v", err)
		return nil, err
	}
	stats.TotalUses = usageCounts.TotalUses
	stats.TotalOrders = usageCounts.TotalOrders

	// 统计总收入和总折扣（需要重新构建查询）
	amountsQuery := r.data.db.WithContext(ctx).Model(&model.CouponUsage{})
	if appID != "" {
		amountsQuery = amountsQuery.Where("coupon_code IN (SELECT coupon_code FROM coupon WHERE app_id = ? AND deleted_at IS NULL)", appID)
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
		CouponCode     string
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
			c.coupon_code,
			c.max_uses,
			COALESCE(COUNT(cu.id), 0) as total_uses,
			COALESCE(COUNT(DISTINCT cu.payment_order_id), 0) as total_orders,
			COALESCE(SUM(cu.final_amount), 0) as total_revenue,
			COALESCE(SUM(cu.discount_amount), 0) as total_discount,
			CASE 
				WHEN c.max_uses > 0 THEN (COALESCE(COUNT(cu.id), 0) * 100.0 / c.max_uses)
				ELSE 0
			END as conversion_rate
		`).
		Joins("LEFT JOIN coupon_usage cu ON c.coupon_code = cu.coupon_code")

	if appID != "" {
		statsQuery = statsQuery.Where("c.app_id = ? AND c.deleted_at IS NULL", appID)
	} else {
		statsQuery = statsQuery.Where("c.deleted_at IS NULL")
	}

	statsQuery = statsQuery.Group("c.coupon_code, c.max_uses").
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
			CouponCode:     sr.CouponCode,
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
