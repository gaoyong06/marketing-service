package data

import (
	"marketing-service/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// Migrate 执行数据库迁移
func Migrate(db *gorm.DB, logger log.Logger) error {
	l := log.NewHelper(log.With(logger, "module", "data/migration"))

	l.Info("starting database migration")

	// 按顺序迁移所有表
	tables := []interface{}{
		&model.Campaign{},
		&model.Audience{},
		&model.Task{},
		&model.Reward{},
		&model.CampaignTask{},
		&model.RewardGrant{},
		&model.InventoryReservation{},
		&model.RedeemCode{},
		&model.TaskCompletionLog{},
	}

	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			l.Errorf("failed to migrate table: %v", err)
			return err
		}
	}

	l.Info("database migration completed successfully")
	return nil
}

// MigrateWithOptions 执行数据库迁移（带选项）
func MigrateWithOptions(db *gorm.DB, logger log.Logger, options ...func(*gorm.DB) *gorm.DB) error {
	l := log.NewHelper(log.With(logger, "module", "data/migration"))

	l.Info("starting database migration with options")

	// 应用选项
	for _, option := range options {
		db = option(db)
	}

	return Migrate(db, logger)
}
