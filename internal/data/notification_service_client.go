package data

import (
	"context"
	"time"

	"marketing-service/internal/conf"
	"marketing-service/internal/biz"
	marketingErrors "marketing-service/internal/errors"

	notificationv1 "xinyuan_tech/notification-service/api/notification/v1"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// notificationClient 实现 biz.NotificationClient 接口
type notificationClient struct {
	client notificationv1.NotificationClient
	conn   *grpc.ClientConn
	log    *log.Helper
}

// NewNotificationClient 创建通知服务客户端
func NewNotificationClient(c *conf.Client, logger log.Logger) (biz.NotificationClient, error) {
	// 检查配置中是否有 notification service 配置
	if c == nil || c.Notification == nil {
		// 如果没有配置，返回空实现
		return &noopNotificationClient{log: log.NewHelper(logger)}, nil
	}

	// 从配置获取 notification service 地址
	grpcAddr := c.Notification.Target
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1:9103" // notification-service 的默认 gRPC 端口
	}

	timeout := 5 * time.Second
	if c.Notification.Timeout != nil {
		timeout = c.Notification.Timeout.AsDuration()
	}

	// 创建 gRPC 连接
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, pkgErrors.WrapErrorWithLang(ctx, err, marketingErrors.ErrCodeNotificationServiceUnavailable)
	}

	client := notificationv1.NewNotificationClient(conn)

	return &notificationClient{
		client: client,
		conn:   conn,
		log:    log.NewHelper(logger),
	}, nil
}

// SendSMS 发送短信
func (c *notificationClient) SendSMS(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	req := &notificationv1.SendRequest{
		UserId:     uint64(userID),
		TemplateId: templateID,
		Channels:   []string{"sms"},
		Params:     params,
		Priority:   3, // 默认优先级
		Async:      true,
	}

	resp, err := c.client.Send(ctx, req)
	if err != nil {
		c.log.Errorf("Failed to send SMS via notification-service: %v", err)
		return pkgErrors.WrapErrorWithLang(ctx, err, marketingErrors.ErrCodeNotificationServiceUnavailable)
	}

	if resp.Status == "failed" {
		c.log.Warnf("SMS send failed: status=%s, message=%s", resp.Status, resp.Message)
		return pkgErrors.NewBizErrorWithLang(ctx, marketingErrors.ErrCodeNotificationSendFailed)
	}

	c.log.Infof("SMS notification sent: id=%s, status=%s", resp.NotificationId, resp.Status)
	return nil
}

// SendEmail 发送邮件
func (c *notificationClient) SendEmail(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	req := &notificationv1.SendRequest{
		UserId:     uint64(userID),
		TemplateId: templateID,
		Channels:   []string{"email"},
		Params:     params,
		Priority:   3, // 默认优先级
		Async:      true,
	}

	resp, err := c.client.Send(ctx, req)
	if err != nil {
		c.log.Errorf("Failed to send email via notification-service: %v", err)
		return pkgErrors.WrapErrorWithLang(ctx, err, marketingErrors.ErrCodeNotificationServiceUnavailable)
	}

	if resp.Status == "failed" {
		c.log.Warnf("Email send failed: status=%s, message=%s", resp.Status, resp.Message)
		return pkgErrors.NewBizErrorWithLang(ctx, marketingErrors.ErrCodeNotificationSendFailed)
	}

	c.log.Infof("Email notification sent: id=%s, status=%s", resp.NotificationId, resp.Status)
	return nil
}

// Close 关闭连接
func (c *notificationClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// noopNotificationClient 空实现，当 notification-service 未启用时使用
type noopNotificationClient struct {
	log *log.Helper
}

func (c *noopNotificationClient) SendSMS(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	c.log.Infof("Notification service disabled, SMS not sent: user_id=%d, template_id=%s", userID, templateID)
	return nil
}

func (c *noopNotificationClient) SendEmail(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	c.log.Infof("Notification service disabled, Email not sent: user_id=%d, template_id=%s", userID, templateID)
	return nil
}

func (c *noopNotificationClient) Close() error {
	return nil
}
