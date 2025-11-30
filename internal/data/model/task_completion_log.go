package model

import (
	"time"

	"gorm.io/datatypes"
)

// TaskCompletionLog 任务完成记录表（事件日志表：记录每次任务完成事件）
type TaskCompletionLog struct {
	CompletionID string         `gorm:"column:completion_id;primaryKey;type:varchar(32);comment:完成记录ID（唯一标识）"`
	TaskID       string         `gorm:"column:task_id;type:varchar(32);not null;index:idx_task_user;comment:任务ID"`
	TaskName     string         `gorm:"column:task_name;type:varchar(128);comment:任务名称（冗余字段，用于报表）"`
	CampaignID   string         `gorm:"column:campaign_id;type:varchar(32);index:idx_campaign_id;comment:活动ID"`
	CampaignName string         `gorm:"column:campaign_name;type:varchar(128);comment:活动名称（冗余字段，用于报表）"`
	UserID       int64          `gorm:"column:user_id;type:bigint;not null;index:idx_task_user,idx_user_id;comment:用户ID"`
	TenantID     string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID        string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	GrantID      string         `gorm:"column:grant_id;type:varchar(32);index:idx_grant_id;comment:关联的奖励授予ID（如果触发了奖励发放）"`
	ProgressData datatypes.JSON `gorm:"column:progress_data;type:json;comment:任务进度数据（JSON格式，如：{\"invited_count\": 3, \"target\": 5}）"`
	TriggerEvent string         `gorm:"column:trigger_event;type:varchar(64);comment:触发事件（如：USER_REGISTER, ORDER_PAID）"`
	CompletedAt  time.Time      `gorm:"column:completed_at;type:datetime;not null;index:idx_completed_at;comment:完成时间"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
}

// TableName 指定表名
func (TaskCompletionLog) TableName() string {
	return "task_completion_log"
}
