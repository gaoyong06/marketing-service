package model

import (
	"time"
)

// InventoryReservation 库存预占表（防止超发，支持高并发）
type InventoryReservation struct {
	ReservationID string    `gorm:"column:reservation_id;primaryKey;type:varchar(32);comment:预占ID（唯一标识）"`
	ResourceID    string    `gorm:"column:resource_id;type:varchar(64);not null;index:idx_resource_status;comment:资源ID（如：优惠券模板ID、商品SKU、总库存Key）"`
	CampaignID    string    `gorm:"column:campaign_id;type:varchar(32);comment:活动ID（可选，用于活动维度的库存控制）"`
	UserID        int64     `gorm:"column:user_id;type:bigint;not null;index:idx_user_id;comment:用户ID"`
	Quantity      int       `gorm:"column:quantity;type:int;not null;default:1;comment:预占数量"`
	Status        string    `gorm:"column:status;type:varchar(16);not null;default:PENDING;index:idx_resource_status;comment:状态：PENDING(预占中)/CONFIRMED(已核销)/CANCELLED(已回滚)/EXPIRED(已过期)"`
	ExpireAt      time.Time `gorm:"column:expire_at;type:datetime;not null;index:idx_expire_at;comment:预占过期时间（超时未核销自动释放）"`
	CreatedAt     time.Time `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt     time.Time `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (InventoryReservation) TableName() string {
	return "inventory_reservation"
}

// InventoryReservationStatus 库存预占状态常量
const (
	InventoryReservationStatusPending   = "PENDING"   // 预占中
	InventoryReservationStatusConfirmed = "CONFIRMED" // 已核销
	InventoryReservationStatusCancelled = "CANCELLED" // 已回滚
	InventoryReservationStatusExpired   = "EXPIRED"   // 已过期
)
