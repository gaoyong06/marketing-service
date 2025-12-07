package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
// 极简重构：仅保留优惠券功能，移除复杂营销活动系统
var ProviderSet = wire.NewSet(
	NewCouponUseCase,
)
