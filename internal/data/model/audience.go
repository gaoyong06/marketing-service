package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Audience 受众表
type Audience struct {
	AudienceID   string         `gorm:"column:audience_id;primaryKey;type:varchar(32);comment:受众ID（唯一标识）"`
	TenantID     string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID        string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	Name         string         `gorm:"column:name;type:varchar(128);not null;comment:受众名称"`
	AudienceType string         `gorm:"column:audience_type;type:varchar(32);not null;index:idx_audience_type;comment:受众类型：TAG/SEGMENT/LIST/ALL"`
	RuleConfig   datatypes.JSON `gorm:"column:rule_config;type:json;not null;comment:圈选规则配置（JSON格式）"`
	Status       string         `gorm:"column:status;type:varchar(16);not null;default:ACTIVE;index:idx_status;comment:状态：ACTIVE/PAUSED/ENDED"`
	Description  string         `gorm:"column:description;type:varchar(512);comment:描述"`
	CreatedBy    string         `gorm:"column:created_by;type:varchar(64);comment:创建人"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (Audience) TableName() string {
	return "audience"
}

// AudienceType 受众类型常量
const (
	AudienceTypeTag     = "TAG"     // 标签圈选
	AudienceTypeSegment = "SEGMENT" // 画像分群
	AudienceTypeList    = "LIST"    // 上传名单
	AudienceTypeAll     = "ALL"     // 全量用户
)
