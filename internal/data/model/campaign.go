package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Campaign 活动表
type Campaign struct {
	CampaignID      string         `gorm:"column:campaign_id;primaryKey;type:varchar(32);comment:活动ID（唯一标识）"`
	TenantID        string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID           string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	CampaignName    string         `gorm:"column:campaign_name;type:varchar(128);not null;comment:活动名称"`
	CampaignType    string         `gorm:"column:campaign_type;type:varchar(32);not null;index:idx_campaign_type;comment:活动类型：REDEEM_CODE/TASK_REWARD/DIRECT_SEND"`
	StartTime       time.Time      `gorm:"column:start_time;type:datetime;not null;index:idx_time_range;comment:开始时间"`
	EndTime         time.Time      `gorm:"column:end_time;type:datetime;not null;index:idx_time_range;comment:结束时间"`
	AudienceConfig  datatypes.JSON `gorm:"column:audience_config;type:json;comment:受众配置（JSON格式，支持多受众组合）"`
	ValidatorConfig datatypes.JSON `gorm:"column:validator_config;type:json;comment:校验规则配置（1:N关系，轻量级组合直接存JSON）"`
	Status          string         `gorm:"column:status;type:varchar(16);not null;default:ACTIVE;index:idx_status;comment:状态：ACTIVE/PAUSED/ENDED"`
	Description     string         `gorm:"column:description;type:varchar(512);comment:活动描述"`
	CreatedBy       string         `gorm:"column:created_by;type:varchar(64);comment:创建人"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (Campaign) TableName() string {
	return "campaign"
}

// CampaignType 活动类型常量
const (
	CampaignTypeRedeemCode = "REDEEM_CODE" // 兑换码活动
	CampaignTypeTaskReward = "TASK_REWARD" // 任务奖励活动
	CampaignTypeDirectSend = "DIRECT_SEND" // 直接发放活动
)

// CampaignStatus 活动状态常量
const (
	CampaignStatusActive = "ACTIVE" // 活动中
	CampaignStatusPaused = "PAUSED" // 已暂停
	CampaignStatusEnded  = "ENDED"  // 已结束
)
