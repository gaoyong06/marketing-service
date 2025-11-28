package biz

import (
	"context"
	"time"
)

// TenantType 租户类型
type TenantType int

const (
	TenantTypePlatform TenantType = iota
	TenantTypeChannel
	TenantTypeEnterprise
)

// Tenant 租户信息
type Tenant struct {
	TenantID       string                 `json:"tenant_id"`
	TenantName     string                 `json:"tenant_name"`
	TenantType     TenantType             `json:"tenant_type"`
	ParentTenantID string                 `json:"parent_tenant_id,omitempty"`
	Status         int32                  `json:"status"`
	QuotaConfig    map[string]interface{} `json:"quota_config,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// TenantRepo 租户仓储接口
type TenantRepo interface {
	GetTenant(ctx context.Context, tenantID string) (*Tenant, error)
	CheckQuota(ctx context.Context, tenantID, quotaType string, count int32) (bool, error)
	ConsumeQuota(ctx context.Context, tenantID, quotaType string, count int32) error
	CheckAndConsumeQuota(ctx context.Context, tenantID, quotaType, limitType string, amount int32, productCode string) (bool, int32, error)
}
