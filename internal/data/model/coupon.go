package model

import (
	"time"

	"gorm.io/gorm"
)

// Coupon 优惠券表
type Coupon struct {
	CouponID      int64          `gorm:"column:coupon_id;primaryKey;autoIncrement;comment:优惠券ID（自增主键）"`
	CouponCode    string         `gorm:"column:coupon_code;primaryKey;type:varchar(50);comment:优惠码（唯一标识）"`
	AppID         string         `gorm:"column:app_id;type:varchar(64);not null;index:idx_app_id;comment:应用ID"`
	DiscountType  string         `gorm:"column:discount_type;type:varchar(16);not null;comment:折扣类型: percent(百分比)/fixed(固定金额)"`
	DiscountValue int64          `gorm:"column:discount_value;type:bigint(20);not null;comment:折扣值(百分比或分)"`
	Currency      string         `gorm:"column:currency;type:enum('CNY','USD','EUR');not null;default:'CNY';comment:货币单位: CNY(人民币)/USD(美元)/EUR(欧元)，仅固定金额类型需要"`
	ValidFrom     time.Time      `gorm:"column:valid_from;type:datetime;not null;index:idx_valid_time;comment:生效时间"`
	ValidUntil    time.Time      `gorm:"column:valid_until;type:datetime;not null;index:idx_valid_time;comment:过期时间"`
	MaxUses       int32          `gorm:"column:max_uses;type:int(11);not null;default:1;comment:最大使用次数"`
	UsedCount     int32          `gorm:"column:used_count;type:int(11);not null;default:0;comment:已使用次数"`
	MinAmount     int64          `gorm:"column:min_amount;type:bigint(20);not null;default:0;comment:最低消费金额(分)"`
	Status        string         `gorm:"column:status;type:enum('active','inactive','expired');not null;default:'active';index:idx_status;comment:优惠券状态: active(激活-可使用)/inactive(禁用-不可使用)/expired(已过期-系统自动标记)"`
	CreatedAt     time.Time      `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (Coupon) TableName() string {
	return "coupon"
}

// CouponUsage 优惠券使用记录表
type CouponUsage struct {
	CouponUsageID  string    `gorm:"column:coupon_usage_id;primaryKey;type:varchar(32);comment:使用记录ID（唯一标识）"`
	CouponCode     string    `gorm:"column:coupon_code;type:varchar(50);not null;index:idx_coupon_code;comment:优惠券码"`
	UID            uint64    `gorm:"column:uid;type:bigint(20);not null;index:idx_uid;comment:用户ID"`
	PaymentOrderID string    `gorm:"column:payment_order_id;type:varchar(64);not null;index:idx_payment_order_id;comment:支付订单ID（payment-service的业务订单号orderId）"`
	PaymentID      string    `gorm:"column:payment_id;type:varchar(64);not null;index:idx_payment_id;comment:支付ID"`
	OriginalAmount int64     `gorm:"column:original_amount;type:bigint(20);not null;comment:原价(分)"`
	DiscountAmount int64     `gorm:"column:discount_amount;type:bigint(20);not null;comment:折扣金额(分)"`
	FinalAmount    int64     `gorm:"column:final_amount;type:bigint(20);not null;comment:实付金额(分)"`
	UsedAt         time.Time `gorm:"column:used_at;type:datetime;not null;default:CURRENT_TIMESTAMP;index:idx_used_at;comment:使用时间"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间"`
}

// TableName 指定表名
func (CouponUsage) TableName() string {
	return "coupon_usage"
}
