package data

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/go-kratos/kratos/v2/log"
	"marketing-service/internal/biz"
)

// TenantModel 租户模型（本地缓存模型，非主数据源）
type TenantModel struct {
	ID             int64      `gorm:"primaryKey;autoIncrement"`
	TenantID       string     `gorm:"type:varchar(32);uniqueIndex;not null"`
	TenantName     string     `gorm:"type:varchar(128);not null"`
	TenantType     string     `gorm:"type:varchar(32);not null"`
	ParentTenantID string     `gorm:"type:varchar(32)"`
	Status         int32      `gorm:"type:tinyint;not null;default:1"`
	QuotaConfig    string     `gorm:"type:text"`
	LastSyncAt     time.Time  `gorm:"type:datetime;not null"`
	CreatedAt      time.Time  `gorm:"type:datetime;not null"`
	UpdatedAt      time.Time  `gorm:"type:datetime;not null"`
	DeletedAt      *time.Time `gorm:"index"`
}

// TableName 设置表名
func (TenantModel) TableName() string {
	return "tenant_cache"
}

type tenantRepo struct {
	data   *Data
	client *TenantServiceClient
	log    *log.Helper
}

// CheckAndConsumeQuota 检查并消费租户配额
func (r *tenantRepo) CheckAndConsumeQuota(ctx context.Context, tenantID, quotaType, limitType string, amount int32, productCode string) (bool, int32, error) {
	// 首先检查配额
	hasQuota, err := r.CheckQuota(ctx, tenantID, quotaType, amount)
	if err != nil {
		r.log.Errorf("Failed to check quota: %v", err)
		return false, 0, fmt.Errorf("failed to check quota: %w", err)
	}

	if !hasQuota {
		r.log.Warnf("Insufficient quota for tenant %s, quota type %s, amount %d", tenantID, quotaType, amount)
		return false, 0, nil
	}

	// 消费配额
	err = r.ConsumeQuota(ctx, tenantID, quotaType, amount)
	if err != nil {
		r.log.Errorf("Failed to consume quota: %v", err)
		return false, 0, fmt.Errorf("failed to consume quota: %w", err)
	}

	// 获取剩余配额
	remaining := int32(1000) // 默认值

	// 尝试从Redis获取已使用配额和限制
	quotaKey := fmt.Sprintf("quota:%s:%s", tenantID, quotaType)
	limitKey := fmt.Sprintf("limit:%s:%s", tenantID, quotaType)

	// 获取已使用配额
	usedStr, err := r.data.redis.Get(ctx, quotaKey).Result()
	if err == nil && usedStr != "" {
		used, err := strconv.ParseInt(usedStr, 10, 64)
		if err == nil {
			// 获取配额限制
			limitStr, err := r.data.redis.Get(ctx, limitKey).Result()
			if err == nil && limitStr != "" {
				limit, err := strconv.ParseInt(limitStr, 10, 64)
				if err == nil {
					remaining = int32(limit - used)
					if remaining < 0 {
						remaining = 0
					}
				}
			}
		}
	}

	r.log.Infof("Successfully consumed quota for tenant %s, quota type %s, amount %d, remaining %d", tenantID, quotaType, amount, remaining)
	return true, remaining, nil
}

// NewTenantRepo 创建租户仓库实现
func NewTenantRepo(data *Data, client *TenantServiceClient, logger log.Logger) biz.TenantRepo {
	return &tenantRepo{
		data:   data,
		client: client,
		log:    log.NewHelper(logger),
	}
}

// GetTenant 获取租户信息
func (r *tenantRepo) GetTenant(ctx context.Context, tenantID string) (*biz.Tenant, error) {
	// 首先尝试从Redis缓存获取
	cacheKey := fmt.Sprintf("tenant:%s", tenantID)
	cachedData, err := r.data.redis.Get(ctx, cacheKey).Result()

	if err == nil {
		// 缓存命中，解析数据
		var tenant biz.Tenant
		if err := json.Unmarshal([]byte(cachedData), &tenant); err == nil {
			return &tenant, nil
		}
		// 解析失败，继续从数据库获取
	}

	// 从本地缓存数据库获取
	var model TenantModel
	if err := r.data.db.Where("tenant_id = ?", tenantID).First(&model).Error; err != nil {
		// 本地缓存未命中，从tenant-service获取
		r.log.Infof("Tenant %s not found in local cache, fetching from tenant service", tenantID)

		// 使用模拟客户端获取租户信息
		tenant, err := r.client.GetTenant(ctx, tenantID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tenant from service: %w", err)
		}

		// 将租户信息保存到本地缓存
		quotaConfigJSON, _ := json.Marshal(tenant.QuotaConfig)
		cacheModel := &TenantModel{
			TenantID:       tenant.TenantID,
			TenantName:     tenant.TenantName,
			TenantType:     convertTenantTypeToString(tenant.TenantType),
			ParentTenantID: tenant.ParentTenantID,
			Status:         tenant.Status,
			QuotaConfig:    string(quotaConfigJSON),
			LastSyncAt:     time.Now(),
			CreatedAt:      tenant.CreatedAt,
			UpdatedAt:      tenant.UpdatedAt,
		}

		// 尝试写入本地缓存
		if err := r.data.db.Create(cacheModel).Error; err != nil {
			r.log.Warnf("Failed to cache tenant %s locally: %v", tenantID, err)
		}

		// 缓存到Redis
		cacheData, err := json.Marshal(tenant)
		if err == nil {
			r.data.redis.Set(ctx, cacheKey, string(cacheData), time.Hour)
		}

		return tenant, nil
	}

	// 解析配额配置
	var quotaConfig map[string]interface{}
	if model.QuotaConfig != "" {
		if err := json.Unmarshal([]byte(model.QuotaConfig), &quotaConfig); err != nil {
			r.log.Warnf("failed to unmarshal quota config for tenant %s: %v", tenantID, err)
		}
	}

	// 构建返回对象
	tenant := &biz.Tenant{
		TenantID:       model.TenantID,
		TenantName:     model.TenantName,
		TenantType:     convertStringToTenantType(model.TenantType),
		ParentTenantID: model.ParentTenantID,
		Status:         model.Status,
		QuotaConfig:    quotaConfig,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}

	// 缓存到Redis，有效期1小时
	cacheData, err := json.Marshal(tenant)
	if err == nil {
		r.data.redis.Set(ctx, cacheKey, string(cacheData), time.Hour)
	}

	return tenant, nil
}

// CheckQuota 检查租户配额
func (r *tenantRepo) CheckQuota(ctx context.Context, tenantID, quotaType string, count int32) (bool, error) {
	// 使用模拟客户端检查配额
	r.log.Infof("Checking quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)

	// 首先尝试使用客户端检查配额
	hasQuota, err := r.client.CheckQuota(ctx, tenantID, quotaType, count)
	if err != nil {
		r.log.Warnf("Failed to check quota from tenant service: %v, falling back to local check", err)
		// 如果调用失败，回退到本地Redis检查
		return r.checkQuotaFromRedis(ctx, tenantID, quotaType, count)
	}

	return hasQuota, nil
}

// checkQuotaFromRedis 从Redis中检查配额
func (r *tenantRepo) checkQuotaFromRedis(ctx context.Context, tenantID, quotaType string, count int32) (bool, error) {
	quotaKey := fmt.Sprintf("quota:%s:%s", tenantID, quotaType)
	limitKey := fmt.Sprintf("limit:%s:%s", tenantID, quotaType)

	// 获取已使用配额
	var used int64 = 0
	usedStr, err := r.data.redis.Get(ctx, quotaKey).Result()
	if err != nil {
		if err != redis.Nil {
			r.log.Warnf("Failed to get quota usage from Redis: %v", err)
		}
		// Redis.Nil 意味着key不存在，将used设为0
	} else if usedStr != "" {
		// 如果有值，尝试解析
		used, err = strconv.ParseInt(usedStr, 10, 64)
		if err != nil {
			r.log.Warnf("Failed to parse quota usage from Redis: %v", err)
			// 解析错误，使用默认值0
		}
	}

	// 获取配额限制
	var limit int64 = 1000 // 默认限制
	limitStr, err := r.data.redis.Get(ctx, limitKey).Result()
	if err != nil {
		if err != redis.Nil {
			r.log.Warnf("Failed to get quota limit from Redis: %v", err)
		}
		// Redis.Nil 意味着key不存在，使用默认限制
	} else if limitStr != "" {
		// 如果有值，尝试解析
		limitVal, err := strconv.ParseInt(limitStr, 10, 64)
		if err == nil {
			limit = limitVal
		} else {
			r.log.Warnf("Failed to parse quota limit from Redis: %v", err)
		}
	}

	// 尝试从租户信息中获取配额设置
	tenantInfo, err := r.GetTenant(ctx, tenantID)
	if err != nil {
		r.log.Warnf("Failed to get tenant info for quota check: %v", err)
		// 获取租户信息失败，使用默认限制检查
		r.log.Infof("Using default quota check: used=%d, limit=%d, count=%d", used, limit, count)
		return used+int64(count) <= limit, nil
	}

	// 检查租户状态
	if tenantInfo.Status != 1 {
		r.log.Warnf("Tenant %s is not active, status: %d", tenantID, tenantInfo.Status)
		return false, nil
	}

	// 平台租户没有配额限制
	if tenantInfo.TenantType == biz.TenantTypePlatform {
		r.log.Infof("Platform tenant %s has unlimited quota", tenantID)
		return true, nil
	}

	// 从租户配置中获取配额限制
	if tenantInfo.QuotaConfig != nil {
		if quotaLimit, ok := tenantInfo.QuotaConfig[quotaType]; ok {
			switch v := quotaLimit.(type) {
			case float64:
				limit = int64(v)
			case int:
				limit = int64(v)
			case int64:
				limit = v
			case string:
				limitVal, err := strconv.ParseInt(v, 10, 64)
				if err == nil {
					limit = limitVal
				} else {
					r.log.Warnf("Failed to parse quota limit from tenant config: %v", err)
				}
			}
		}
	}

	r.log.Infof("Quota check for tenant %s: used=%d, limit=%d, count=%d, result=%v",
		tenantID, used, limit, count, used+int64(count) <= limit)
	return used+int64(count) <= limit, nil
}

// ConsumeQuota 消费租户配额
func (r *tenantRepo) ConsumeQuota(ctx context.Context, tenantID, quotaType string, count int32) error {
	if count <= 0 {
		r.log.Warnf("Invalid quota consumption request: count must be positive, got %d", count)
		return fmt.Errorf("invalid quota consumption: count must be positive")
	}

	r.log.Infof("Consuming quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)

	// 首先检查是否有足够配额
	hasQuota, err := r.CheckQuota(ctx, tenantID, quotaType, count)
	if err != nil {
		r.log.Errorf("Failed to check quota before consumption: %v", err)
		return fmt.Errorf("failed to check quota before consumption: %w", err)
	}

	if !hasQuota {
		r.log.Warnf("Insufficient quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)
		return fmt.Errorf("insufficient quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)
	}

	// 使用客户端消费配额
	err = r.client.ConsumeQuota(ctx, tenantID, quotaType, count)
	if err != nil {
		r.log.Errorf("Failed to consume quota from tenant service: %v", err)
		return fmt.Errorf("failed to consume quota from tenant service: %w", err)
	}

	// 更新Redis缓存
	quotaKey := fmt.Sprintf("quota:%s:%s", tenantID, quotaType)
	_, err = r.data.redis.IncrBy(ctx, quotaKey, int64(count)).Result()
	if err != nil {
		// Redis错误不应该阻止消费成功，只记录日志
		r.log.Warnf("Failed to update quota usage in Redis: %v", err)
	} else {
		// 设置过期时间（如果不存在）
		// 默认设置30天过期时间
		expireErr := r.data.redis.Expire(ctx, quotaKey, 30*24*time.Hour).Err()
		if expireErr != nil {
			r.log.Warnf("Failed to set expiration for quota key %s: %v", quotaKey, expireErr)
		}
	}

	r.log.Infof("Successfully consumed quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)
	return nil
}

// 转换租户类型字符串为枚举值
func convertStringToTenantType(typeStr string) biz.TenantType {
	switch typeStr {
	case "PLATFORM":
		return biz.TenantTypePlatform
	case "CHANNEL":
		return biz.TenantTypeChannel
	case "ENTERPRISE":
		return biz.TenantTypeEnterprise
	default:
		return biz.TenantTypePlatform
	}
}

// 转换租户类型枚举值为字符串
func convertTenantTypeToString(tenantType biz.TenantType) string {
	switch tenantType {
	case biz.TenantTypePlatform:
		return "platform"
	case biz.TenantTypeChannel:
		return "channel"
	case biz.TenantTypeEnterprise:
		return "enterprise"
	default:
		return "unknown"
	}
}
