package data

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis/v8"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"marketing-service/internal/conf"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedis,
	NewCampaignRepo,
	NewRedeemCodeRepo,
	NewTenantRepo,
	NewTenantServiceClient,
)

// Data ..
type Data struct {
	db    *gorm.DB
	redis *redis.Client
}

// GormWriter 自定义 GORM 日志写入器
type GormWriter struct {
	helper *log.Helper
}

// Printf 实现 logger.Writer 接口
func (w *GormWriter) Printf(format string, args ...interface{}) {
	w.helper.Infof(format, args...)
}

// NewDB creates a new database connection.
func NewDB(conf *conf.Data, l log.Logger) *gorm.DB {
	logHelper := log.NewHelper(l)

	// 创建 GORM 日志配置
	writer := &GormWriter{helper: logHelper}
	gormLogger := logger.New(
		writer,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(mysql.Open(conf.Database.Source), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})

	if err != nil {
		logHelper.Fatalf("failed opening connection to mysql: %v", err)
	}

	// 设置连接池参数
	sqlDB, err := db.DB()
	if err != nil {
		logHelper.Fatalf("failed to get database: %v", err)
	}

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(int(conf.Database.MaxOpenConns))
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(int(conf.Database.MaxIdleConns))
	// 设置连接最大生命周期
	sqlDB.SetConnMaxLifetime(conf.Database.ConnMaxLifetime.AsDuration())

	logHelper.Info("mysql connected")

	return db
}

// NewRedis creates a new redis client.
func NewRedis(conf *conf.Data, l log.Logger) *redis.Client {
	logHelper := log.NewHelper(l)

	client := redis.NewClient(&redis.Options{
		Addr:         conf.Redis.Addr,
		Password:     conf.Redis.Password,
		DB:           int(conf.Redis.Db),
		DialTimeout:  conf.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  conf.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: conf.Redis.WriteTimeout.AsDuration(),
		PoolSize:     int(conf.Redis.PoolSize),
		MinIdleConns: int(conf.Redis.MinIdleConns),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logHelper.Fatalf("failed to connect redis: %v", err)
	}

	logHelper.Info("redis connected")
	return client
}

// NewData .
func NewData(db *gorm.DB, redis *redis.Client, l log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(l)
	logHelper.Info("creating data resources")

	d := &Data{
		db:    db,
		redis: redis,
	}

	return d, func() {
		logHelper.Info("closing data resources")
		if redis != nil {
			if err := redis.Close(); err != nil {
				logHelper.Errorf("redis close error: %v", err)
			}
		}
		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				logHelper.Errorf("db connection error: %v", err)
				return
			}
			if err := sqlDB.Close(); err != nil {
				logHelper.Errorf("db close error: %v", err)
			}
		}
	}, nil
}
