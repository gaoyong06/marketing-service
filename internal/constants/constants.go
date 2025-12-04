package constants

// ========== 通用状态常量 ==========

// Status 通用状态
const (
	StatusActive  = "ACTIVE"  // 激活/活动中
	StatusPaused  = "PAUSED"  // 已暂停
	StatusEnded   = "ENDED"   // 已结束
	StatusPending = "PENDING" // 待处理
)

// ========== Campaign 相关常量 ==========

// CampaignType 活动类型
const (
	CampaignTypeRedeemCode = "REDEEM_CODE" // 兑换码活动
	CampaignTypeTaskReward = "TASK_REWARD" // 任务奖励活动
	CampaignTypeDirectSend = "DIRECT_SEND" // 直接发放活动
)

// CampaignStatus 活动状态
const (
	CampaignStatusActive = "ACTIVE" // 活动中
	CampaignStatusPaused = "PAUSED" // 已暂停
	CampaignStatusEnded  = "ENDED"  // 已结束
)

// ========== Reward 相关常量 ==========

// RewardType 奖励类型
const (
	RewardTypeCoupon       = "COUPON"       // 优惠券
	RewardTypePoints       = "POINTS"       // 积分
	RewardTypeRedeemCode   = "REDEEM_CODE"  // 兑换码
	RewardTypeSubscription = "SUBSCRIPTION" // 订阅/会员
)

// RewardStatus 奖励状态
const (
	RewardStatusActive = "ACTIVE" // 激活
	RewardStatusPaused = "PAUSED" // 已暂停
	RewardStatusEnded  = "ENDED"  // 已结束
)

// ========== Task 相关常量 ==========

// TaskType 任务类型
const (
	TaskTypeInvite   = "INVITE"   // 邀请好友
	TaskTypePurchase = "PURCHASE" // 购买商品
	TaskTypeShare    = "SHARE"    // 分享内容
	TaskTypeSignIn   = "SIGN_IN"  // 签到打卡
)

// TaskStatus 任务状态
const (
	TaskStatusActive = "ACTIVE" // 激活
	TaskStatusPaused = "PAUSED" // 已暂停
	TaskStatusEnded  = "ENDED"  // 已结束
)

// ========== Audience 相关常量 ==========

// AudienceType 受众类型
const (
	AudienceTypeTag     = "TAG"     // 标签圈选
	AudienceTypeSegment = "SEGMENT" // 画像分群
	AudienceTypeList    = "LIST"    // 上传名单
	AudienceTypeAll     = "ALL"     // 全量用户
)

// AudienceStatus 受众状态
const (
	AudienceStatusActive = "ACTIVE" // 激活
	AudienceStatusPaused = "PAUSED" // 已暂停
	AudienceStatusEnded  = "ENDED"  // 已结束
)

// ========== RedeemCode 相关常量 ==========

// RedeemCodeStatus 兑换码状态
const (
	RedeemCodeStatusActive   = "ACTIVE"   // 可用
	RedeemCodeStatusRedeemed = "REDEEMED" // 已兑换
	RedeemCodeStatusExpired  = "EXPIRED"  // 已过期
	RedeemCodeStatusRevoked  = "REVOKED"  // 已作废
)

// CodeType 码类型
const (
	CodeTypeCoupon   = "COUPON"   // 优惠券
	CodeTypeDiscount = "DISCOUNT" // 折扣码
	CodeTypeGift     = "GIFT"     // 礼品码
)

// ========== RewardGrant 相关常量 ==========

// RewardGrantStatus 奖励发放状态
const (
	RewardGrantStatusPending     = "PENDING"     // 待处理
	RewardGrantStatusGenerated   = "GENERATED"   // 已生成
	RewardGrantStatusReserved    = "RESERVED"    // 已预占
	RewardGrantStatusDistributed = "DISTRIBUTED" // 已发放
	RewardGrantStatusUsed        = "USED"        // 已使用
	RewardGrantStatusExpired     = "EXPIRED"     // 已过期
)

// ========== InventoryReservation 相关常量 ==========

// InventoryReservationStatus 库存预占状态
const (
	InventoryReservationStatusPending   = "PENDING"   // 预占中
	InventoryReservationStatusConfirmed = "CONFIRMED" // 已确认
	InventoryReservationStatusCancelled = "CANCELLED" // 已取消
	InventoryReservationStatusExpired   = "EXPIRED"   // 已过期
)

// ========== Generator 相关常量 ==========

// GeneratorType 生成器类型
const (
	GeneratorTypeCode   = "CODE"   // 兑换码生成器
	GeneratorTypeCoupon = "COUPON" // 优惠券生成器
	GeneratorTypePoints = "POINTS" // 积分生成器
)

// ========== Discount 相关常量 ==========

// DiscountType 折扣类型
const (
	DiscountTypeAmount  = "AMOUNT"  // 金额折扣
	DiscountTypePercent = "PERCENT" // 百分比折扣
)

// ========== Event 相关常量 ==========

// EventType 事件类型
const (
	EventTypeUserRegister = "USER_REGISTER" // 用户注册
	EventTypeOrderPaid    = "ORDER_PAID"    // 订单支付
	EventTypeUserSignIn   = "USER_SIGN_IN"  // 用户签到
)

// ========== RocketMQ 相关常量 ==========

// RocketMQTopic RocketMQ Topic 名称
const (
	RocketMQTopicTaskCompleted = "marketing.task.completed" // 任务完成事件 Topic（默认值）
)
