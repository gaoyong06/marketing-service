package biz

import (
	"context"
	"time"

	"marketing-service/internal/constants"

	"github.com/gaoyong06/go-pkg/errors"

	"github.com/go-kratos/kratos/v2/log"
)

// Coupon 优惠券领域对象
type Coupon struct {
	CouponID      int64     // 优惠券ID（自增主键）
	CouponCode    string    // 优惠码（业务唯一标识）
	AppID         string    // 应用ID
	DiscountType  string    // 折扣类型
	DiscountValue int64     // 折扣值
	Currency      string    // 货币单位: CNY, USD, EUR 等，仅固定金额类型需要
	ValidFrom     time.Time // 生效时间
	ValidUntil    time.Time // 过期时间
	MaxUses       int32     // 最大使用次数
	UsedCount     int32     // 已使用次数
	MinAmount     int64     // 最低消费金额
	Status        string    // 状态
	CreatedAt     time.Time // 创建时间
	UpdatedAt     time.Time // 更新时间
}

// CouponUsage 优惠券使用记录领域对象
type CouponUsage struct {
	CouponUsageID  string
	CouponCode     string
	AppID          string // 应用ID
	UID            uint64
	PaymentOrderID string // 支付订单ID（payment-service的业务订单号orderId）
	PaymentID      string
	OriginalAmount int64
	DiscountAmount int64
	FinalAmount    int64
	UsedAt         time.Time // 使用时间
	CreatedAt      time.Time // 创建时间
}

// CouponRepo 优惠券仓储接口
type CouponRepo interface {
	Save(context.Context, *Coupon) (*Coupon, error)
	Update(context.Context, *Coupon) (*Coupon, error)
	FindByCode(context.Context, string) (*Coupon, error)
	List(context.Context, string, string, int, int) ([]*Coupon, int64, error) // appID, status, page, pageSize
	Delete(context.Context, string) error
	IncrementUsedCount(context.Context, string) error // 原子性增加使用次数
	CreateUsage(context.Context, *CouponUsage) error
	UseCoupon(context.Context, string, string, uint64, string, string, int64, int64, int64) error // 使用优惠券（事务操作）：code, appID, userID, paymentOrderID, paymentID, originalAmount, discountAmount, finalAmount
	ListUsages(context.Context, string, int, int) ([]*CouponUsage, int64, error)          // couponCode, page, pageSize
	GetStats(context.Context, string) (*CouponStats, error)
	GetSummaryStats(context.Context, string) (*SummaryStats, error) // appID（可选），获取汇总统计
}

// CouponStats 优惠券统计信息
type CouponStats struct {
	CouponCode     string  // 优惠码
	TotalUses      int32   // 总使用次数
	TotalOrders    int32   // 总订单数
	TotalRevenue   int64   // 总营收
	TotalDiscount  int64   // 总折扣
	ConversionRate float32 // 转化率
}

// SummaryStats 汇总统计信息
type SummaryStats struct {
	TotalCoupons          int32
	ActiveCoupons         int32
	TotalUses             int32
	TotalOrders           int32
	TotalRevenue          int64
	TotalDiscount         int64
	AverageConversionRate float32
	TopCoupons            []*CouponStats // 前N个优惠券的详细统计
}

// CouponUseCase 优惠券用例
type CouponUseCase struct {
	repo CouponRepo
	log  *log.Helper
}

// NewCouponUseCase 创建优惠券用例
func NewCouponUseCase(repo CouponRepo, logger log.Logger) *CouponUseCase {
	return &CouponUseCase{
		repo: repo,
		log:  log.NewHelper(logger),
	}
}

// Create 创建优惠券
func (uc *CouponUseCase) Create(ctx context.Context, c *Coupon) (*Coupon, error) {
	if c.Status == "" {
		c.Status = constants.CouponStatusActive
	}
	// 确保创建时 UsedCount 为 0
	if c.UsedCount == 0 {
		c.UsedCount = 0
	}
	// 如果货币单位为空，设置默认值为 CNY
	if c.Currency == "" {
		c.Currency = constants.CouponCurrencyCNY
	}
	// 验证货币单位是否有效（数据库 enum 会再次验证，但这里可以提前发现问题）
	if !isValidCurrency(c.Currency) {
		return nil, errors.NewBizError(errors.ErrCodeInvalidArgument, "zh-CN")
	}
	now := time.Now()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	if c.UpdatedAt.IsZero() {
		c.UpdatedAt = now
	}
	return uc.repo.Save(ctx, c)
}

// isValidCurrency 验证货币单位是否有效
func isValidCurrency(currency string) bool {
	for _, validCurrency := range constants.ValidCouponCurrencies {
		if currency == validCurrency {
			return true
		}
	}
	return false
}

// Get 获取优惠券
func (uc *CouponUseCase) Get(ctx context.Context, code string) (*Coupon, error) {
	return uc.repo.FindByCode(ctx, code)
}

// List 列出优惠券
func (uc *CouponUseCase) List(ctx context.Context, appID, status string, page, pageSize int) ([]*Coupon, int64, error) {
	return uc.repo.List(ctx, appID, status, page, pageSize)
}

// Update 更新优惠券
func (uc *CouponUseCase) Update(ctx context.Context, c *Coupon) (*Coupon, error) {
	// 业务规则验证
	if !c.ValidFrom.IsZero() && !c.ValidUntil.IsZero() && !c.ValidFrom.Before(c.ValidUntil) {
		return nil, errors.NewBizError(errors.ErrCodeBusinessRuleViolation, "zh-CN")
	}
	if c.DiscountValue <= 0 {
		return nil, errors.NewBizError(errors.ErrCodeBusinessRuleViolation, "zh-CN")
	}
	if c.DiscountType == constants.CouponDiscountTypePercent && c.DiscountValue > 100 {
		return nil, errors.NewBizError(errors.ErrCodeBusinessRuleViolation, "zh-CN")
	}
	// 验证货币单位是否有效（如果提供了货币单位）
	if c.Currency != "" && !isValidCurrency(c.Currency) {
		return nil, errors.NewBizError(errors.ErrCodeInvalidArgument, "zh-CN")
	}

	c.UpdatedAt = time.Now()
	return uc.repo.Update(ctx, c)
}

// Delete 删除优惠券
func (uc *CouponUseCase) Delete(ctx context.Context, code string) error {
	return uc.repo.Delete(ctx, code)
}

// Validate 验证优惠券（供 Payment Service 调用）
func (uc *CouponUseCase) Validate(ctx context.Context, code, appID string, amount int64) (*Coupon, int64, error) {
	coupon, err := uc.repo.FindByCode(ctx, code)
	if err != nil {
		return nil, 0, err
	}
	if coupon == nil {
		return nil, 0, nil
	}

	// 检查应用ID
	if coupon.AppID != appID {
		return nil, 0, nil
	}

	// 检查状态
	if coupon.Status != constants.CouponStatusActive {
		return nil, 0, nil
	}

	// 检查有效期
	now := time.Now()
	if now.Before(coupon.ValidFrom) || now.After(coupon.ValidUntil) {
		return nil, 0, nil
	}

	// 检查使用次数（MaxUses = 0 表示无限制）
	if coupon.MaxUses > 0 && coupon.UsedCount >= coupon.MaxUses {
		return nil, 0, nil
	}

	// 检查最低消费金额
	if amount < coupon.MinAmount {
		return nil, 0, nil
	}

	// 计算折扣金额
	var discountAmount int64
	if coupon.DiscountType == constants.CouponDiscountTypePercent {
		discountAmount = amount * coupon.DiscountValue / 100
	} else {
		discountAmount = coupon.DiscountValue
		// 折扣金额不能超过订单金额
		if discountAmount > amount {
			discountAmount = amount
		}
	}

	return coupon, discountAmount, nil
}

// Use 使用优惠券（供 Payment Service 调用）
// 注意：需要在事务中执行，确保数据一致性
// paymentOrderID: payment-service的业务订单号orderId
func (uc *CouponUseCase) Use(ctx context.Context, code, appID string, userID uint64, paymentOrderID, paymentID string, originalAmount, discountAmount, finalAmount int64) error {
	// 使用事务确保原子性：先增加使用次数，再创建使用记录
	// 如果创建使用记录失败，需要回滚使用次数的增加
	// 注意：这里依赖 Repository 层的事务支持，如果 Repository 不支持事务，需要在 UseCase 层实现
	return uc.repo.UseCoupon(ctx, code, appID, userID, paymentOrderID, paymentID, originalAmount, discountAmount, finalAmount)
}

// GetStats 获取优惠券统计
func (uc *CouponUseCase) GetStats(ctx context.Context, code string) (*CouponStats, error) {
	return uc.repo.GetStats(ctx, code)
}

// ListUsages 列出优惠券使用记录
func (uc *CouponUseCase) ListUsages(ctx context.Context, code string, page, pageSize int) ([]*CouponUsage, int64, error) {
	return uc.repo.ListUsages(ctx, code, page, pageSize)
}

// GetSummaryStats 获取汇总统计
func (uc *CouponUseCase) GetSummaryStats(ctx context.Context, appID string) (*SummaryStats, error) {
	return uc.repo.GetSummaryStats(ctx, appID)
}
