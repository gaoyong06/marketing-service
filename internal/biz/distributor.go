package biz

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"marketing-service/internal/errors"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// Distributor 发放器接口
type Distributor interface {
	Distribute(ctx context.Context, req *DistributionRequest) error
}

// DistributionRequest 发放请求
type DistributionRequest struct {
	GrantID    string
	RewardID   string
	RewardType string
	UserID     int64
	Content    string                 // 奖励内容（JSON）
	Config     map[string]interface{} // 发放器配置
}

// DistributorService 发放器服务
type DistributorService struct {
	distributors map[string]Distributor
	log          *log.Helper
}

// NewDistributorService 创建发放器服务
func NewDistributorService(notificationService *NotificationService, logger log.Logger) *DistributorService {
	ds := &DistributorService{
		distributors: make(map[string]Distributor),
		log:          log.NewHelper(logger),
	}

	// 注册内置发放器
	ds.Register("AUTO", NewAutoDistributor())
	ds.Register("WEBHOOK", NewWebhookDistributor())

	// 创建并注册邮件发放器
	emailDistributor := NewEmailDistributor().(*EmailDistributor)
	emailDistributor.SetNotificationService(notificationService)
	ds.Register("EMAIL", emailDistributor)

	// 创建并注册短信发放器
	smsDistributor := NewSMSDistributor().(*SMSDistributor)
	smsDistributor.SetNotificationService(notificationService)
	ds.Register("SMS", smsDistributor)

	return ds
}

// Register 注册发放器
func (ds *DistributorService) Register(distributorType string, distributor Distributor) {
	ds.distributors[distributorType] = distributor
}

// Distribute 执行发放
func (ds *DistributorService) Distribute(ctx context.Context, req *DistributionRequest) error {
	if req.Config == nil {
		// 无配置，使用自动发放
		distributor := ds.distributors["AUTO"]
		if distributor == nil {
			return pkgErrors.NewBizErrorWithLang(ctx, errors.ErrCodeDistributorNotFound)
		}
		return distributor.Distribute(ctx, req)
	}

	distributorType, ok := req.Config["type"].(string)
	if !ok {
		distributorType = "AUTO" // 默认自动发放
	}

	distributor, exists := ds.distributors[distributorType]
	if !exists {
		ds.log.Warnf("unknown distributor type: %s, using auto", distributorType)
		distributor = ds.distributors["AUTO"]
		if distributor == nil {
			return pkgErrors.NewBizErrorWithLang(ctx, errors.ErrCodeDistributorNotFound)
		}
	}

	return distributor.Distribute(ctx, req)
}

// ========== 内置发放器实现 ==========

// AutoDistributor 自动发放器
type AutoDistributor struct {
	log *log.Helper
}

// NewAutoDistributor 创建自动发放器
func NewAutoDistributor() Distributor {
	return &AutoDistributor{}
}

// Distribute 自动发放（直接完成，无需额外操作）
func (d *AutoDistributor) Distribute(ctx context.Context, req *DistributionRequest) error {
	// 自动发放器不需要额外操作，奖励已经记录在 RewardGrant 中
	// 实际使用时，可以通过消息队列或事件总线通知其他系统
	return nil
}

// WebhookDistributor Webhook 发放器
type WebhookDistributor struct {
	httpClient *http.Client
	log        *log.Helper
}

// NewWebhookDistributor 创建 Webhook 发放器
func NewWebhookDistributor() Distributor {
	return &WebhookDistributor{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Distribute 通过 Webhook 发放
func (d *WebhookDistributor) Distribute(ctx context.Context, req *DistributionRequest) error {
	webhookURL, ok := req.Config["webhook_url"].(string)
	if !ok || webhookURL == "" {
		return pkgErrors.NewBizErrorWithLang(ctx, errors.ErrCodeWebhookURLNotConfigured)
	}

	// 构建请求体
	payload := map[string]interface{}{
		"grant_id":    req.GrantID,
		"reward_id":   req.RewardID,
		"reward_type": req.RewardType,
		"user_id":     req.UserID,
		"content":     req.Content,
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// 发送 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewReader(payloadJSON))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := d.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return pkgErrors.NewBizErrorWithLang(ctx, errors.ErrCodeWebhookRequestFailed)
	}

	return nil
}

// EmailDistributor 邮件发放器
type EmailDistributor struct {
	notificationService *NotificationService
	log                 *log.Helper
}

// NewEmailDistributor 创建邮件发放器
func NewEmailDistributor() Distributor {
	return &EmailDistributor{}
}

// SetNotificationService 设置通知服务（用于依赖注入）
func (d *EmailDistributor) SetNotificationService(ns *NotificationService) {
	d.notificationService = ns
}

// Distribute 通过邮件发放
func (d *EmailDistributor) Distribute(ctx context.Context, req *DistributionRequest) error {
	if d.notificationService == nil {
		// 降级处理：记录日志
		if d.log != nil {
			d.log.Infof("notification service not available, logging email send to user %d", req.UserID)
		}
		return nil
	}

	// 从配置中获取模板ID
	templateID, ok := req.Config["template_id"].(string)
	if !ok || templateID == "" {
		templateID = "reward_email" // 默认模板
	}

	// 构建模板参数
	params := map[string]string{
		"grant_id":    req.GrantID,
		"reward_id":   req.RewardID,
		"reward_type": req.RewardType,
		"content":     req.Content,
	}

	// 发送邮件
	return d.notificationService.SendEmail(ctx, req.UserID, templateID, params)
}

// SMSDistributor 短信发放器
type SMSDistributor struct {
	notificationService *NotificationService
	log                 *log.Helper
}

// NewSMSDistributor 创建短信发放器
func NewSMSDistributor() Distributor {
	return &SMSDistributor{}
}

// SetNotificationService 设置通知服务（用于依赖注入）
func (d *SMSDistributor) SetNotificationService(ns *NotificationService) {
	d.notificationService = ns
}

// Distribute 通过短信发放
func (d *SMSDistributor) Distribute(ctx context.Context, req *DistributionRequest) error {
	if d.notificationService == nil {
		// 降级处理：记录日志
		if d.log != nil {
			d.log.Infof("notification service not available, logging SMS send to user %d", req.UserID)
		}
		return nil
	}

	// 从配置中获取模板ID
	templateID, ok := req.Config["template_id"].(string)
	if !ok || templateID == "" {
		templateID = "reward_sms" // 默认模板
	}

	// 构建模板参数
	params := map[string]string{
		"grant_id":    req.GrantID,
		"reward_id":   req.RewardID,
		"reward_type": req.RewardType,
		"content":     req.Content,
	}

	// 发送短信
	return d.notificationService.SendSMS(ctx, req.UserID, templateID, params)
}

// ParseDistributorConfig 解析发放器配置
func ParseDistributorConfig(configJSON string) (map[string]interface{}, error) {
	if configJSON == "" {
		return nil, nil
	}

	var config map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, err
	}

	return config, nil
}
