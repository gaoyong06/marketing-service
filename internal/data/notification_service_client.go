package data

import (
	"context"
	"fmt"
	"time"

	"marketing-service/conf"

	notificationv1 "xinyuan_tech/notification-service/api/notification/v1"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NotificationClient 定义通知服务客户端接口
// 注意：这个接口与 biz.NotificationClient 相同，用于实现 biz 层接口
type NotificationClient interface {
	SendSMS(ctx context.Context, userID int64, templateID string, params map[string]string) error
	SendEmail(ctx context.Context, userID int64, templateID string, params map[string]string) error
	Close() error
}

// notificationClient 实现 NotificationClient 接口
type notificationClient struct {
	client notificationv1.NotificationClient
	conn   *grpc.ClientConn
	log    *log.Helper
}

// NewNotificationClient 创建通知服务客户端
func NewNotificationClient(c *conf.Bootstrap, logger log.Logger) (NotificationClient, error) {
	// 检查配置中是否有 notification service 配置
	if c.Client == nil || c.Client.NotificationService == nil {
		// 如果没有配置，返回空实现
		return &noopNotificationClient{log: log.NewHelper(logger)}, nil
	}

	// 从配置获取 notification service 地址
	grpcAddr := c.Client.NotificationService.Target
	if grpcAddr == "" {
		grpcAddr = "127.0.0.1:9103" // notification-service 的默认 gRPC 端口
	}

	timeout := 5 * time.Second
	if c.Client.NotificationService.Timeout != nil {
		timeout = c.Client.NotificationService.Timeout.AsDuration()
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
		return nil, fmt.Errorf("failed to connect to notification service: %w", err)
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
		return fmt.Errorf("failed to call notification service: %w", err)
	}

	if resp.Status == "failed" {
		c.log.Warnf("SMS send failed: status=%s, message=%s", resp.Status, resp.Message)
		return fmt.Errorf("notification send failed: %s", resp.Message)
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
		return fmt.Errorf("failed to call notification service: %w", err)
	}

	if resp.Status == "failed" {
		c.log.Warnf("Email send failed: status=%s, message=%s", resp.Status, resp.Message)
		return fmt.Errorf("notification send failed: %s", resp.Message)
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
