package model

// Coupon 优惠券表
type Coupon struct {
	Code          string `gorm:"column:code;primaryKey;type:varchar(50);comment:优惠码（唯一标识）"`
	AppID         string `gorm:"column:app_id;type:varchar(32);not null;index:idx_app_id;comment:应用ID"`
	DiscountType  string `gorm:"column:discount_type;type:varchar(16);not null;comment:折扣类型: percent(百分比)/fixed(固定金额)"`
	DiscountValue int64  `gorm:"column:discount_value;type:bigint(20);not null;comment:折扣值(百分比或分)"`
	ValidFrom     int64  `gorm:"column:valid_from;type:bigint(20);not null;index:idx_valid_time;comment:生效时间(timestamp)"`
	ValidUntil    int64  `gorm:"column:valid_until;type:bigint(20);not null;index:idx_valid_time;comment:过期时间(timestamp)"`
	MaxUses       int32  `gorm:"column:max_uses;type:int(11);not null;default:1;comment:最大使用次数"`
	UsedCount     int32  `gorm:"column:used_count;type:int(11);not null;default:0;comment:已使用次数"`
	MinAmount     int64  `gorm:"column:min_amount;type:bigint(20);not null;default:0;comment:最低消费金额(分)"`
	Status        string `gorm:"column:status;type:varchar(16);not null;default:active;index:idx_status;comment:状态: active/inactive/expired"`
	CreatedAt     int64  `gorm:"column:created_at;type:bigint(20);not null;comment:创建时间(timestamp)"`
	UpdatedAt     int64  `gorm:"column:updated_at;type:bigint(20);not null;comment:更新时间(timestamp)"`
}

// TableName 指定表名
func (Coupon) TableName() string {
	return "coupon"
}

// CouponUsage 优惠券使用记录表
type CouponUsage struct {
	ID             string `gorm:"column:id;primaryKey;type:varchar(32);comment:使用记录ID（唯一标识）"`
	CouponCode     string `gorm:"column:coupon_code;type:varchar(50);not null;index:idx_coupon_code;comment:优惠券码"`
	UserID         uint64 `gorm:"column:user_id;type:bigint(20);not null;index:idx_user_id;comment:用户ID"`
	OrderID        string `gorm:"column:order_id;type:varchar(64);not null;index:idx_order_id;comment:订单ID"`
	PaymentID      string `gorm:"column:payment_id;type:varchar(64);not null;index:idx_payment_id;comment:支付ID"`
	OriginalAmount int64  `gorm:"column:original_amount;type:bigint(20);not null;comment:原价(分)"`
	DiscountAmount int64  `gorm:"column:discount_amount;type:bigint(20);not null;comment:折扣金额(分)"`
	FinalAmount    int64  `gorm:"column:final_amount;type:bigint(20);not null;comment:实付金额(分)"`
	UsedAt         int64  `gorm:"column:used_at;type:bigint(20);not null;index:idx_used_at;comment:使用时间(timestamp)"`
	CreatedAt      int64  `gorm:"column:created_at;type:bigint(20);not null;comment:创建时间(timestamp)"`
}

// TableName 指定表名
func (CouponUsage) TableName() string {
	return "coupon_usage"
}
