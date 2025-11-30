package model

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Task 任务表
type Task struct {
	TaskID          string         `gorm:"column:task_id;primaryKey;type:varchar(32);comment:任务ID（唯一标识）"`
	TenantID        string         `gorm:"column:tenant_id;type:varchar(32);not null;index:idx_tenant_app;comment:租户ID"`
	AppID           string         `gorm:"column:app_id;type:varchar(32);not null;index:idx_tenant_app;comment:应用ID"`
	Name            string         `gorm:"column:name;type:varchar(128);not null;comment:任务名称"`
	TaskType        string         `gorm:"column:task_type;type:varchar(32);not null;index:idx_task_type;comment:任务类型：INVITE/PURCHASE/SHARE/SIGN_IN"`
	TriggerConfig   datatypes.JSON `gorm:"column:trigger_config;type:json;comment:触发配置（JSON格式：Event, Condition）"`
	ConditionConfig datatypes.JSON `gorm:"column:condition_config;type:json;not null;comment:完成条件配置（JSON格式）"`
	RewardID        string         `gorm:"column:reward_id;type:varchar(32);index:idx_reward_id;comment:关联奖励ID（可选）"`
	Status          string         `gorm:"column:status;type:varchar(16);not null;default:ACTIVE;index:idx_status;comment:状态：ACTIVE/PAUSED/ENDED"`
	StartTime       time.Time      `gorm:"column:start_time;type:datetime;not null;index:idx_time_range;comment:开始时间"`
	EndTime         time.Time      `gorm:"column:end_time;type:datetime;not null;index:idx_time_range;comment:结束时间"`
	MaxCount        int            `gorm:"column:max_count;type:int;not null;default:0;comment:最大完成次数（0表示无限制）"`
	Description     string         `gorm:"column:description;type:varchar(512);comment:描述"`
	CreatedBy       string         `gorm:"column:created_by;type:varchar(64);comment:创建人"`
	CreatedAt       time.Time      `gorm:"column:created_at;type:datetime;not null;autoCreateTime;comment:创建时间"`
	UpdatedAt       time.Time      `gorm:"column:updated_at;type:datetime;not null;autoUpdateTime;comment:更新时间"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index:idx_deleted_at;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (Task) TableName() string {
	return "task"
}

// TaskType 任务类型常量
const (
	TaskTypeInvite   = "INVITE"   // 邀请好友
	TaskTypePurchase = "PURCHASE" // 购买商品
	TaskTypeShare    = "SHARE"    // 分享内容
	TaskTypeSignIn   = "SIGN_IN"  // 签到打卡
)
