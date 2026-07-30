package service

import (
	"context"
	"io"
	"testing"

	"marketing-service/internal/biz"
	_ "marketing-service/internal/errors"

	"github.com/gaoyong06/go-pkg/middleware/app_id"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
	v1 "marketing-service/api/marketing_service/v1"
)

type couponRepoStub struct {
	coupon  *biz.Coupon
	deleted bool
}

func (stub *couponRepoStub) Save(context.Context, *biz.Coupon) (*biz.Coupon, error) {
	return stub.coupon, nil
}
func (stub *couponRepoStub) Update(context.Context, *biz.Coupon) (*biz.Coupon, error) {
	return stub.coupon, nil
}
func (stub *couponRepoStub) FindByCode(context.Context, string) (*biz.Coupon, error) {
	return stub.coupon, nil
}
func (stub *couponRepoStub) List(context.Context, string, string, int, int) ([]*biz.Coupon, int64, error) {
	return nil, 0, nil
}
func (stub *couponRepoStub) Delete(context.Context, string) error                { stub.deleted = true; return nil }
func (stub *couponRepoStub) IncrementUsedCount(context.Context, string) error    { return nil }
func (stub *couponRepoStub) CreateUsage(context.Context, *biz.CouponUsage) error { return nil }
func (stub *couponRepoStub) UseCoupon(context.Context, string, string, string, string, string, int64, int64, int64) error {
	return nil
}
func (stub *couponRepoStub) ListUsages(context.Context, string, int, int) ([]*biz.CouponUsage, int64, error) {
	return nil, 0, nil
}
func (stub *couponRepoStub) GetStats(context.Context, string) (*biz.CouponStats, error) {
	return &biz.CouponStats{}, nil
}
func (stub *couponRepoStub) GetSummaryStats(context.Context, string) (*biz.SummaryStats, error) {
	return &biz.SummaryStats{}, nil
}

func TestGetCouponRejectsDifferentApp(t *testing.T) {
	service, _ := newMarketingServiceForTest(&biz.Coupon{CouponCode: "WELCOME", AppID: "app-2"})
	ctx := app_id.WithAppID(context.Background(), "app-1")
	if _, err := service.GetCoupon(ctx, &v1.GetCouponRequest{CouponCode: "WELCOME"}); err == nil {
		t.Fatal("expected coupon app ownership error")
	}
}

func TestDeleteCouponRejectsDifferentAppWithoutMutation(t *testing.T) {
	service, repo := newMarketingServiceForTest(&biz.Coupon{CouponCode: "WELCOME", AppID: "app-2"})
	ctx := app_id.WithAppID(context.Background(), "app-1")
	var reply *emptypb.Empty
	reply, err := service.DeleteCoupon(ctx, &v1.DeleteCouponRequest{CouponCode: "WELCOME"})
	if err == nil || reply != nil {
		t.Fatal("expected coupon app ownership error")
	}
	if repo.deleted {
		t.Fatal("coupon must not be deleted")
	}
}

func newMarketingServiceForTest(coupon *biz.Coupon) (*MarketingService, *couponRepoStub) {
	repo := &couponRepoStub{coupon: coupon}
	logger := log.NewStdLogger(io.Discard)
	return NewMarketingService(biz.NewCouponUseCase(repo, logger), logger), repo
}
