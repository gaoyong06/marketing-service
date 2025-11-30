package model

import (
	"time"

	"gorm.io/datatypes"
)

// RewardGrant 奖励发放表（核心业务数据表）
type RewardGrant struct {
	GrantID         string         `gorm:"column:grant_id;primaryKey;type:varchar(32);comment:授予ID（唯一标识）"`
	RewardID        string         `gorm:"column:reward_id;type:varchar(32);not null;index:idx_reward_id,idx_reward_version;comment:奖励模板ID"`
	RewardName      string         `gorm:"column:reward_name;type:varchar(128);comment:奖励名称（冗余字段，用于列表展示）"`
	RewardType      string         `gorm:"column:reward_type;type:varchar(32);not null;index:idx_reward_type;comment:奖励类型（冗余字段，用于筛选）"`
	RewardVersion   int            `gorm:"column:reward_version;type:int;not null;index:idx_reward_version;comment:奖励版本号"`
	ContentSnapshot datatypes.JSON `gorm:"column:content_snapshot;type:json;not null;comment:奖励内容快照"`
	GeneratorConfig datatypes.JSON `gorm:"column:generator_config;type:json;comment:生成配置快照（JSON格式，替代generator_id）"`
	CampaignID      string         `gorm:"column:campaign_id;type:varchar(32);index:idx_campaign_id;comment:活动ID"`
	CampaignName    string         `gorm:"column:campaign_name;type:varchar(128);comment:活动名称（冗余字段，避免JOIN）"`
	TaskID          string         `gorm:"column:task_id;type:varchar(32);comment:任务ID"`
	TaskName        string         `gorm:"column:task_name;type:varchar(128);comment:任务名称（冗余字段，避免JOIN）"`
	TenantID        string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID           string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	UserID          int64          `gorm:"column:user_id;type:bigint;index:idx_user_status;comment:用户ID"`
	Status          string         `gorm:"column:status;type:varchar(16);not null;default:PENDING;index:idx_status,idx_user_status;comment:状态：PENDING/GENERATED/RESERVED/DISTRIBUTED/USED/EXPIRED"`
	ReservedAt      *time.Time     `gorm:"column:reserved_at;type:datetime;comment:预占时间"`
	DistributedAt   *time.Time     `gorm:"column:distributed_at;type:datetime;comment:发放时间"`
	UsedAt          *time.Time     `gorm:"column:used_at;type:datetime;comment:使用时间"`
	ExpireTime      *time.Time     `gorm:"column:expire_time;type:datetime;index:idx_expire_time;comment:过期时间"`
	ErrorMessage    string         `gorm:"column:error_message;type:varchar(512);comment:错误信息（发放失败时记录）"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (RewardGrant) TableName() string {
	return "reward_grant"
}

// RewardGrantStatus 奖励发放状态常量
const (
	RewardGrantStatusPending     = "PENDING"     // 待处理
	RewardGrantStatusGenerated   = "GENERATED"   // 已生成
	RewardGrantStatusReserved    = "RESERVED"    // 已预占
	RewardGrantStatusDistributed = "DISTRIBUTED" // 已发放
	RewardGrantStatusUsed        = "USED"        // 已使用
	RewardGrantStatusExpired     = "EXPIRED"     // 已过期
)
