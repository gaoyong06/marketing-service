package service

import (
	"context"
	"time"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/biz"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/gaoyong06/go-pkg/middleware/app_id"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
)

// MarketingService 营销服务（极简重构版：仅保留优惠券功能）
type MarketingService struct {
	v1.UnimplementedMarketingServer

	cuc *biz.CouponUseCase
	log *log.Helper
}

// NewMarketingService 创建营销服务（极简重构版）
func NewMarketingService(
	cuc *biz.CouponUseCase,
	logger log.Logger,
) *MarketingService {
	return &MarketingService{
		cuc: cuc,
		log: log.NewHelper(log.With(logger, "module", "service/marketing")),
	}
}

// ========== Coupon Management API ==========

// CreateCoupon 创建优惠券
func (s *MarketingService) CreateCoupon(ctx context.Context, req *v1.CreateCouponRequest) (*v1.CreateCouponReply, error) {
	// 获取 appId（只从 Context，由中间件从 Header 提取）
	appID := app_id.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, pkgErrors.NewBizErrorWithLang(ctx, pkgErrors.ErrCodeInvalidArgument)
	}

	coupon := &biz.Coupon{
		CouponCode:    req.CouponCode,
		AppID:         appID,
		DiscountType:  req.DiscountType,
		DiscountValue: req.DiscountValue,
		Currency:      req.Currency, // 货币单位，如果为空则 biz 层会设置默认值 CNY
		ValidFrom:     time.Unix(req.ValidFrom, 0),
		ValidUntil:    time.Unix(req.ValidUntil, 0),
		MaxUses:       req.MaxUses,
		MinAmount:     req.MinAmount,
	}

	result, err := s.cuc.Create(ctx, coupon)
	if err != nil {
		s.log.Errorf("failed to create coupon: %v", err)
		return nil, err
	}

	return &v1.CreateCouponReply{
		Coupon: s.toProtoCoupon(result),
	}, nil
}

// GetCoupon 获取优惠券
func (s *MarketingService) GetCoupon(ctx context.Context, req *v1.GetCouponRequest) (*v1.GetCouponReply, error) {
	coupon, err := s.cuc.Get(ctx, req.CouponCode)
	if err != nil {
		s.log.Errorf("failed to get coupon: %v", err)
		return nil, err
	}
	if coupon == nil {
		return nil, pkgErrors.NewBizError(pkgErrors.ErrCodeNotFound, "zh-CN")
	}

	return &v1.GetCouponReply{
		Coupon: s.toProtoCoupon(coupon),
	}, nil
}

// ListCoupons 列出优惠券
func (s *MarketingService) ListCoupons(ctx context.Context, req *v1.ListCouponsRequest) (*v1.ListCouponsReply, error) {
	// 获取 appId（只从 Context，由中间件从 Header 提取）
	appID := app_id.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, pkgErrors.NewBizErrorWithLang(ctx, pkgErrors.ErrCodeInvalidArgument)
	}

	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	coupons, total, err := s.cuc.List(ctx, appID, req.Status, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list coupons: %v", err)
		return nil, err
	}

	protoCoupons := make([]*v1.Coupon, 0, len(coupons))
	for _, c := range coupons {
		protoCoupons = append(protoCoupons, s.toProtoCoupon(c))
	}

	return &v1.ListCouponsReply{
		Coupons:  protoCoupons,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// UpdateCoupon 更新优惠券
func (s *MarketingService) UpdateCoupon(ctx context.Context, req *v1.UpdateCouponRequest) (*v1.UpdateCouponReply, error) {
	coupon, err := s.cuc.Get(ctx, req.CouponCode)
	if err != nil {
		s.log.Errorf("failed to get coupon: %v", err)
		return nil, err
	}
	if coupon == nil {
		return nil, pkgErrors.NewBizError(pkgErrors.ErrCodeNotFound, "zh-CN")
	}

	// 更新字段
	if req.DiscountType != "" {
		coupon.DiscountType = req.DiscountType
	}
	if req.DiscountValue > 0 {
		coupon.DiscountValue = req.DiscountValue
	}
	if req.Currency != "" {
		coupon.Currency = req.Currency
	}
	if req.ValidFrom > 0 {
		coupon.ValidFrom = time.Unix(req.ValidFrom, 0)
	}
	if req.ValidUntil > 0 {
		coupon.ValidUntil = time.Unix(req.ValidUntil, 0)
	}
	if req.MaxUses > 0 {
		coupon.MaxUses = req.MaxUses
	}
	if req.MinAmount >= 0 {
		coupon.MinAmount = req.MinAmount
	}
	if req.Status != "" {
		coupon.Status = req.Status
	}

	result, err := s.cuc.Update(ctx, coupon)
	if err != nil {
		s.log.Errorf("failed to update coupon: %v", err)
		return nil, err
	}

	return &v1.UpdateCouponReply{
		Coupon: s.toProtoCoupon(result),
	}, nil
}

// DeleteCoupon 删除优惠券
func (s *MarketingService) DeleteCoupon(ctx context.Context, req *v1.DeleteCouponRequest) (*emptypb.Empty, error) {
	err := s.cuc.Delete(ctx, req.CouponCode)
	if err != nil {
		s.log.Errorf("failed to delete coupon: %v", err)
		return nil, err
	}

	// 成功时返回 Empty，统一响应格式中的 success 字段会表示操作成功
	return &emptypb.Empty{}, nil
}

// ValidateCoupon 验证优惠券（供 Payment Service 调用）
func (s *MarketingService) ValidateCoupon(ctx context.Context, req *v1.ValidateCouponRequest) (*v1.ValidateCouponReply, error) {
	// 获取 appId（只从 Context，由中间件从 Header 提取）
	appID := app_id.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, pkgErrors.NewBizErrorWithLang(ctx, pkgErrors.ErrCodeInvalidArgument)
	}

	coupon, discountAmount, err := s.cuc.Validate(ctx, req.CouponCode, appID, req.Amount)
	if err != nil {
		s.log.Errorf("failed to validate coupon: %v", err)
		return nil, err
	}

	if coupon == nil {
		return &v1.ValidateCouponReply{
			Valid:   false,
			Message: "优惠券无效或不可用",
		}, nil
	}

	finalAmount := req.Amount - discountAmount
	if finalAmount < 0 {
		finalAmount = 0
	}

	return &v1.ValidateCouponReply{
		Valid:          true,
		Message:        "优惠券有效",
		DiscountAmount: discountAmount,
		FinalAmount:    finalAmount,
		Coupon:         s.toProtoCoupon(coupon),
	}, nil
}

// UseCoupon 使用优惠券（供 Payment Service 调用）
func (s *MarketingService) UseCoupon(ctx context.Context, req *v1.UseCouponRequest) (*v1.UseCouponReply, error) {
	// 获取 appId（从 Context，由中间件从 Header 提取）
	appID := app_id.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, pkgErrors.NewBizErrorWithLang(ctx, pkgErrors.ErrCodeInvalidArgument)
	}

	err := s.cuc.Use(ctx, req.CouponCode, appID, req.UserId, req.PaymentOrderId, req.PaymentId, req.OriginalAmount, req.DiscountAmount, req.FinalAmount)
	if err != nil {
		s.log.Errorf("failed to use coupon: %v", err)
		return &v1.UseCouponReply{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &v1.UseCouponReply{
		Success: true,
		Message: "优惠券使用成功",
	}, nil
}

// GetCouponStats 获取优惠券统计
func (s *MarketingService) GetCouponStats(ctx context.Context, req *v1.GetCouponStatsRequest) (*v1.GetCouponStatsReply, error) {
	stats, err := s.cuc.GetStats(ctx, req.CouponCode)
	if err != nil {
		s.log.Errorf("failed to get coupon stats: %v", err)
		return nil, err
	}

	return &v1.GetCouponStatsReply{
		CouponCode:     stats.CouponCode,
		TotalUses:      stats.TotalUses,
		TotalOrders:    stats.TotalOrders,
		TotalRevenue:   stats.TotalRevenue,
		TotalDiscount:  stats.TotalDiscount,
		ConversionRate: float32(stats.ConversionRate),
	}, nil
}

// ListCouponUsages 列出优惠券使用记录
func (s *MarketingService) ListCouponUsages(ctx context.Context, req *v1.ListCouponUsagesRequest) (*v1.ListCouponUsagesReply, error) {
	page := int(req.Page)
	if page <= 0 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	usages, total, err := s.cuc.ListUsages(ctx, req.CouponCode, page, pageSize)
	if err != nil {
		s.log.Errorf("failed to list coupon usages: %v", err)
		return nil, err
	}

	protoUsages := make([]*v1.CouponUsage, 0, len(usages))
	for _, u := range usages {
		protoUsages = append(protoUsages, s.toProtoCouponUsage(u))
	}

	return &v1.ListCouponUsagesReply{
		Usages:   protoUsages,
		Total:    int32(total),
		Page:     int32(page),
		PageSize: int32(pageSize),
	}, nil
}

// GetCouponsSummaryStats 获取所有优惠券汇总统计
func (s *MarketingService) GetCouponsSummaryStats(ctx context.Context, req *v1.GetCouponsSummaryStatsRequest) (*v1.GetCouponsSummaryStatsReply, error) {
	// 获取 appId（只从 Context，由中间件从 Header 提取）
	appID := app_id.GetAppIDFromContext(ctx)
	if appID == "" {
		return nil, pkgErrors.NewBizErrorWithLang(ctx, pkgErrors.ErrCodeInvalidArgument)
	}

	stats, err := s.cuc.GetSummaryStats(ctx, appID)
	if err != nil {
		s.log.Errorf("failed to get coupons summary stats: %v", err)
		return nil, err
	}

	protoTopCoupons := make([]*v1.CouponStats, 0, len(stats.TopCoupons))
	for _, cs := range stats.TopCoupons {
		protoTopCoupons = append(protoTopCoupons, &v1.CouponStats{
			CouponCode:     cs.CouponCode,
			TotalUses:      cs.TotalUses,
			TotalOrders:    cs.TotalOrders,
			TotalRevenue:   cs.TotalRevenue,
			TotalDiscount:  cs.TotalDiscount,
			ConversionRate: cs.ConversionRate,
		})
	}

	return &v1.GetCouponsSummaryStatsReply{
		TotalCoupons:          stats.TotalCoupons,
		ActiveCoupons:         stats.ActiveCoupons,
		TotalUses:             stats.TotalUses,
		TotalOrders:           stats.TotalOrders,
		TotalRevenue:          stats.TotalRevenue,
		TotalDiscount:         stats.TotalDiscount,
		AverageConversionRate: stats.AverageConversionRate,
		TopCoupons:            protoTopCoupons,
	}, nil
}

// toProtoCoupon 转换为 Proto Coupon
func (s *MarketingService) toProtoCoupon(c *biz.Coupon) *v1.Coupon {
	var validFrom, validUntil, createdAt, updatedAt int64
	if !c.ValidFrom.IsZero() {
		validFrom = c.ValidFrom.Unix()
	}
	if !c.ValidUntil.IsZero() {
		validUntil = c.ValidUntil.Unix()
	}
	if !c.CreatedAt.IsZero() {
		createdAt = c.CreatedAt.Unix()
	}
	if !c.UpdatedAt.IsZero() {
		updatedAt = c.UpdatedAt.Unix()
	}
	return &v1.Coupon{
		CouponCode:    c.CouponCode,
		AppId:         c.AppID,
		DiscountType:  c.DiscountType,
		DiscountValue: c.DiscountValue,
		Currency:      c.Currency,
		ValidFrom:     validFrom,
		ValidUntil:    validUntil,
		MaxUses:       c.MaxUses,
		UsedCount:     c.UsedCount,
		MinAmount:     c.MinAmount,
		Status:        c.Status,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}
}

// toProtoCouponUsage 转换为 Proto CouponUsage
func (s *MarketingService) toProtoCouponUsage(u *biz.CouponUsage) *v1.CouponUsage {
	var usedAt int64
	if !u.UsedAt.IsZero() {
		usedAt = u.UsedAt.Unix()
	}
	return &v1.CouponUsage{
		CouponUsageId:  u.CouponUsageID,
		CouponCode:     u.CouponCode,
		UserId:         u.UID,
		PaymentOrderId: u.PaymentOrderID,
		PaymentId:      u.PaymentID,
		OriginalAmount: u.OriginalAmount,
		DiscountAmount: u.DiscountAmount,
		FinalAmount:    u.FinalAmount,
		UsedAt:         usedAt,
	}
}
