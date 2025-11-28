package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"marketing-service/internal/biz"
	"marketing-service/internal/conf"
)

// TenantServiceClient is a client for the tenant service.
type TenantServiceClient struct {
	conf *conf.Client
	log  *log.Helper
}

// NewTenantServiceClient creates a new tenant service client.
func NewTenantServiceClient(conf *conf.Client, logger log.Logger) *TenantServiceClient {
	return &TenantServiceClient{
		conf: conf,
		log:  log.NewHelper(logger),
	}
}

// GetTenant gets a tenant from the tenant service.
func (c *TenantServiceClient) GetTenant(ctx context.Context, tenantID string) (*biz.Tenant, error) {
	// TODO: Implement actual gRPC client call to tenant service
	// For now, we'll return a mock tenant
	c.log.Infof("Getting tenant %s from tenant service", tenantID)

	// Mock tenant data
	tenantType := biz.TenantTypePlatform
	if len(tenantID) > 3 {
		prefix := tenantID[:3]
		switch prefix {
		case "CH_":
			tenantType = biz.TenantTypeChannel
		case "EN_":
			tenantType = biz.TenantTypeEnterprise
		}
	}

	return &biz.Tenant{
		TenantID:   tenantID,
		TenantName: fmt.Sprintf("Tenant %s", tenantID),
		TenantType: tenantType,
		Status:     1, // Active
		QuotaConfig: map[string]interface{}{
			"campaign":    100,
			"redeem_code": 10000,
			"batch":       50,
		},
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now(),
	}, nil
}

// CheckQuota checks if a tenant has enough quota for a specific operation.
func (c *TenantServiceClient) CheckQuota(ctx context.Context, tenantID, quotaType string, count int32) (bool, error) {
	// TODO: Implement actual gRPC client call to tenant service
	// For now, we'll return a mock result
	c.log.Infof("Checking quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)

	// Mock quota check
	var limit int32
	switch quotaType {
	case "campaign":
		limit = 100
	case "redeem_code":
		limit = 10000
	case "batch":
		limit = 50
	default:
		limit = 1000
	}

	// Simulate a random current usage between 0 and 70% of the limit
	currentUsage := int32(float64(limit) * 0.7)
	remaining := limit - currentUsage

	return count <= remaining, nil
}

// ConsumeQuota consumes quota for a tenant.
func (c *TenantServiceClient) ConsumeQuota(ctx context.Context, tenantID, quotaType string, count int32) error {
	// TODO: Implement actual gRPC client call to tenant service
	// For now, we'll just log the consumption
	c.log.Infof("Consuming quota for tenant %s, quota type %s, count %d", tenantID, quotaType, count)
	return nil
}
