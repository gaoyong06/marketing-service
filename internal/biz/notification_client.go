package biz

import (
	"context"

	"marketing-service/internal/errors"

	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	"github.com/go-kratos/kratos/v2/log"
)

// NotificationClient 通知服务客户端接口
type NotificationClient interface {
	// SendEmail 发送邮件
	SendEmail(ctx context.Context, userID int64, templateID string, params map[string]string) error
	// SendSMS 发送短信
	SendSMS(ctx context.Context, userID int64, templateID string, params map[string]string) error
}

// NotificationService 通知服务（用于集成 notification-service）
type NotificationService struct {
	client NotificationClient
	log    *log.Helper
}

// NewNotificationService 创建通知服务
func NewNotificationService(client NotificationClient, logger log.Logger) *NotificationService {
	return &NotificationService{
		client: client,
		log:    log.NewHelper(logger),
	}
}

// SendEmail 发送邮件通知
func (ns *NotificationService) SendEmail(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	if ns.client == nil {
		ns.log.Warn("notification client is nil, skipping email send")
		return nil
	}

	if err := ns.client.SendEmail(ctx, userID, templateID, params); err != nil {
		ns.log.Errorf("failed to send email to user %d: %v", userID, err)
		return pkgErrors.WrapErrorWithLang(ctx, err, errors.ErrCodeNotificationSendFailed)
	}

	ns.log.Infof("email sent to user %d with template %s", userID, templateID)
	return nil
}

// SendSMS 发送短信通知
func (ns *NotificationService) SendSMS(ctx context.Context, userID int64, templateID string, params map[string]string) error {
	if ns.client == nil {
		ns.log.Warn("notification client is nil, skipping SMS send")
		return nil
	}

	if err := ns.client.SendSMS(ctx, userID, templateID, params); err != nil {
		ns.log.Errorf("failed to send SMS to user %d: %v", userID, err)
		return pkgErrors.WrapErrorWithLang(ctx, err, errors.ErrCodeNotificationSendFailed)
	}

	ns.log.Infof("SMS sent to user %d with template %s", userID, templateID)
	return nil
}
