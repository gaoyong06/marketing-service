package data

import (
	"context"
	"time"

	"marketing-service/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is data providers.
// 极简重构：仅保留优惠券功能，移除复杂营销活动系统
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedis,
	NewCouponRepo,
	// 移除 RocketMQ、Cache、Notification 等复杂依赖
	// 如需缓存，可在 CouponRepo 中直接使用 Redis
)

// Data .
type Data struct {
	db  *gorm.DB
	rdb *redis.Client
	log *log.Helper
}

// NewData .
func NewData(db *gorm.DB, rdb *redis.Client, logger log.Logger) (*Data, func(), error) {
	l := log.NewHelper(log.With(logger, "module", "data"))

	cleanup := func() {
		l.Info("closing the data resources")
		if sqlDB, err := db.DB(); err == nil {
			sqlDB.Close()
		}
		if rdb != nil {
			rdb.Close()
		}
	}

	return &Data{
		db:  db,
		rdb: rdb,
		log: l,
	}, cleanup, nil
}

// NewDB .
func NewDB(c *conf.Data, logger log.Logger) *gorm.DB {
	l := log.NewHelper(log.With(logger, "module", "data/db"))

	db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
		Logger: NewGormLogger(l),
	})
	if err != nil {
		l.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		l.Fatalf("failed to get database instance: %v", err)
	}

	// 设置连接池
	sqlDB.SetMaxIdleConns(int(c.Database.MaxIdleConns))
	sqlDB.SetMaxOpenConns(int(c.Database.MaxOpenConns))
	sqlDB.SetConnMaxLifetime(c.Database.ConnMaxLifetime.AsDuration())

	l.Info("database connected successfully")
	return db
}

// NewRedis .
func NewRedis(c *conf.Data, logger log.Logger) *redis.Client {
	l := log.NewHelper(log.With(logger, "module", "data/redis"))

	rdb := redis.NewClient(&redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
		PoolSize:     int(c.Redis.PoolSize),
		MinIdleConns: int(c.Redis.MinIdleConns),
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		l.Fatalf("failed to connect redis: %v", err)
	}

	l.Info("redis connected successfully")
	return rdb
}

// GormLogger 适配器
type GormLogger struct {
	*log.Helper
}

// NewGormLogger 创建 GORM Logger 适配器
func NewGormLogger(helper *log.Helper) logger.Interface {
	return &GormLogger{Helper: helper}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Helper.Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	if err != nil {
		l.Helper.Errorf("sql error: %v, elapsed: %v, sql: %s, rows: %d", err, elapsed, sql, rows)
	} else {
		l.Helper.Debugf("sql trace: elapsed: %v, sql: %s, rows: %d", elapsed, sql, rows)
	}
}
