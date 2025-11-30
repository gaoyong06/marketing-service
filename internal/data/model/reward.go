package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Reward 奖励表（奖励模板）
type Reward struct {
	RewardID          string         `gorm:"column:reward_id;primaryKey;type:varchar(32);comment:奖励ID（唯一标识）"`
	TenantID          string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID             string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	RewardType        string         `gorm:"column:reward_type;type:varchar(32);not null;index:idx_reward_type;comment:奖励类型：COUPON/POINTS/REDEEM_CODE/SUBSCRIPTION"`
	Name              string         `gorm:"column:name;type:varchar(128);comment:奖励名称"`
	ContentConfig     datatypes.JSON `gorm:"column:content_config;type:json;not null;comment:奖励内容配置（JSON格式）"`
	GeneratorConfig   datatypes.JSON `gorm:"column:generator_config;type:json;comment:生成配置（JSON格式，替代Generator表）"`
	DistributorConfig datatypes.JSON `gorm:"column:distributor_config;type:json;comment:发放配置（JSON格式，替代Distributor表）"`
	ValidatorConfig   datatypes.JSON `gorm:"column:validator_config;type:json;comment:校验规则配置（1:N关系，轻量级组合直接存JSON）"`
	Version           int            `gorm:"column:version;type:int;not null;default:1;index:idx_version;comment:版本号（每次修改时递增）"`
	ValidDays         int            `gorm:"column:valid_days;type:int;not null;default:0;comment:有效期（天数）"`
	ExtraConfig       datatypes.JSON `gorm:"column:extra_config;type:json;comment:额外配置（JSON格式）"`
	Status            string         `gorm:"column:status;type:varchar(16);not null;default:ACTIVE;index:idx_status;comment:状态：ACTIVE/PAUSED/ENDED"`
	Description       string         `gorm:"column:description;type:varchar(512);comment:描述"`
	CreatedBy         string         `gorm:"column:created_by;type:varchar(64);comment:创建人"`
	CreatedAt         time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt         time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (Reward) TableName() string {
	return "reward"
}

// RewardType 奖励类型常量
const (
	RewardTypeCoupon       = "COUPON"       // 优惠券
	RewardTypePoints       = "POINTS"       // 积分
	RewardTypeRedeemCode   = "REDEEM_CODE"  // 兑换码
	RewardTypeSubscription = "SUBSCRIPTION" // 订阅/会员
)
