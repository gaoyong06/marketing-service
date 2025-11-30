package model

import (
	"time"

	"gorm.io/datatypes"
)

// CampaignTask 活动组合任务关系表
type CampaignTask struct {
	CampaignTaskID int64          `gorm:"column:campaign_task_id;primaryKey;autoIncrement;comment:自增主键"`
	CampaignID     string         `gorm:"column:campaign_id;type:varchar(32);not null;uniqueIndex:uk_campaign_task;comment:活动ID"`
	TaskID         string         `gorm:"column:task_id;type:varchar(32);not null;uniqueIndex:uk_campaign_task;index:idx_task_id;comment:任务ID"`
	Config         datatypes.JSON `gorm:"column:config;type:json;comment:组合配置（如：覆盖任务的默认阈值）"`
	SortOrder      int            `gorm:"column:sort_order;type:int;not null;default:0;comment:排序顺序"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (CampaignTask) TableName() string {
	return "campaign_task"
}
