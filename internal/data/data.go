package data

import (
	"context"
	"time"

	"marketing-service/conf"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewDB,
	NewRedis,
	NewRocketMQProducer,
	NewCacheService,
	NewCampaignRepo,
	NewRewardRepo,
	NewRewardGrantRepo,
	NewTaskRepo,
	NewAudienceRepo,
	NewRedeemCodeRepo,
	NewInventoryReservationRepo,
	NewTaskCompletionLogRepo,
	NewCampaignTaskRepo,
	NewNotificationClient,
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

// NewRocketMQProducer 创建 RocketMQ Producer
// 如果配置未启用或配置为空，返回 nil
// 如果连接失败，记录错误日志但返回 nil，不导致服务启动失败
func NewRocketMQProducer(c *conf.Data, logger log.Logger) (rocketmq.Producer, func(), error) {
	l := log.NewHelper(log.With(logger, "module", "data/rocketmq"))

	// 检查是否启用 RocketMQ
	if c.Rocketmq == nil || !c.Rocketmq.Enabled {
		l.Info("RocketMQ is not enabled, skipping producer initialization")
		return nil, func() {}, nil
	}

	// 检查 NameServer 配置
	if len(c.Rocketmq.NameServers) == 0 {
		l.Warn("RocketMQ name_servers is empty, skipping producer initialization")
		return nil, func() {}, nil
	}

	// 创建 Producer
	rmqProducer, err := rocketmq.NewProducer(
		producer.WithNameServer(c.Rocketmq.NameServers),
		producer.WithGroupName(c.Rocketmq.GroupName),
		producer.WithRetry(int(c.Rocketmq.RetryTimes)),
		producer.WithSendMsgTimeout(c.Rocketmq.SendTimeout.AsDuration()),
	)
	if err != nil {
		// 连接失败时记录错误日志，但不导致服务启动失败
		l.Errorf("failed to create RocketMQ producer: %v, name_servers: %v, group: %s", err, c.Rocketmq.NameServers, c.Rocketmq.GroupName)
		return nil, func() {}, nil
	}

	// 启动 Producer
	if err := rmqProducer.Start(); err != nil {
		// 启动失败时记录错误日志，但不导致服务启动失败
		l.Errorf("failed to start RocketMQ producer: %v, name_servers: %v, group: %s", err, c.Rocketmq.NameServers, c.Rocketmq.GroupName)
		// 尝试关闭已创建的 Producer
		_ = rmqProducer.Shutdown()
		return nil, func() {}, nil
	}

	l.Infof("RocketMQ producer started successfully, name_servers: %v, group: %s", c.Rocketmq.NameServers, c.Rocketmq.GroupName)

	cleanup := func() {
		l.Info("closing RocketMQ producer")
		if err := rmqProducer.Shutdown(); err != nil {
			l.Errorf("failed to shutdown RocketMQ producer: %v", err)
		}
	}

	return rmqProducer, cleanup, nil
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
