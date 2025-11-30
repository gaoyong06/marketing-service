package model

import (
	"time"
)

// RedeemCode 兑换码表（简化设计：只负责码的生命周期管理）
type RedeemCode struct {
	Code         string     `gorm:"column:code;primaryKey;type:varchar(64);comment:兑换码（唯一标识）"`
	TenantID     string     `gorm:"column:tenant_id;primaryKey;type:varchar(32);comment:租户ID"`
	AppID        string     `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	GrantID      string     `gorm:"column:grant_id;type:varchar(32);index:idx_grant_id;comment:关联的奖励授予ID（兑换后关联）"`
	CampaignID   string     `gorm:"column:campaign_id;type:varchar(32);index:idx_campaign_reward;comment:所属活动ID"`
	CampaignName string     `gorm:"column:campaign_name;type:varchar(128);comment:活动名称（冗余字段，用于展示）"`
	RewardID     string     `gorm:"column:reward_id;type:varchar(32);not null;index:idx_campaign_reward;comment:关联奖励模板ID"`
	RewardName   string     `gorm:"column:reward_name;type:varchar(128);comment:奖励名称（冗余字段，用于展示）"`
	BatchID      string     `gorm:"column:batch_id;type:varchar(32);index:idx_batch_id;comment:批次ID（批量生成时使用）"`
	Status       string     `gorm:"column:status;type:varchar(16);not null;default:ACTIVE;index:idx_status;comment:状态：ACTIVE(可用)/REDEEMED(已兑换)/EXPIRED(已过期)/REVOKED(已作废)"`
	OwnerUserID  *int64     `gorm:"column:owner_user_id;type:bigint;index:idx_owner_user;comment:拥有者用户ID（预分配场景）"`
	RedeemedBy   *int64     `gorm:"column:redeemed_by;type:bigint;comment:兑换者用户ID（可能与owner不同）"`
	RedeemedAt   *time.Time `gorm:"column:redeemed_at;type:datetime;comment:兑换时间"`
	ExpireAt     *time.Time `gorm:"column:expire_at;type:datetime;index:idx_expire_at;comment:过期时间"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (RedeemCode) TableName() string {
	return "redeem_code"
}

// RedeemCodeStatus 兑换码状态常量
const (
	RedeemCodeStatusActive   = "ACTIVE"   // 可用
	RedeemCodeStatusRedeemed = "REDEEMED" // 已兑换
	RedeemCodeStatusExpired  = "EXPIRED"  // 已过期
	RedeemCodeStatusRevoked  = "REVOKED"  // 已作废
)
