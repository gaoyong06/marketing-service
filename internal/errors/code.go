package errors

import (
	pkgErrors "github.com/gaoyong06/go-pkg/errors"
	i18nPkg "github.com/gaoyong06/go-pkg/middleware/i18n"
)

func init() {
	// 初始化全局错误管理器（使用项目特定的配置）
	pkgErrors.InitGlobalErrorManager("i18n", i18nPkg.Language)
}

// Marketing Service 错误码定义
// 错误码格式：SSMMEE (6位数字)
//   SS: 服务标识，Marketing 固定为 12
//   MM: 模块标识，按业务划分
//   EE: 模块内错误序号
//
// 模块划分：
//   00: 通用模块（复用 go-pkg 通用错误码）
//   01: 活动模块
//   02: 兑换码模块
//   03: 奖励模块
//   04: 任务模块
//   05: 受众模块
//   06: 通知模块
//   07: 分发器模块
//   08-99: 预留扩展

// 活动模块错误码 (120100-120199)
const (
	// ErrCodeCampaignNotFound 活动不存在
	ErrCodeCampaignNotFound = 120101
	// ErrCodeCampaignCreateFailed 活动创建失败
	ErrCodeCampaignCreateFailed = 120102
	// ErrCodeCampaignUpdateFailed 活动更新失败
	ErrCodeCampaignUpdateFailed = 120103
	// ErrCodeCampaignDeleteFailed 活动删除失败
	ErrCodeCampaignDeleteFailed = 120104
)

// 兑换码模块错误码 (120200-120299)
const (
	// ErrCodeRedeemCodeNotFound 兑换码不存在
	ErrCodeRedeemCodeNotFound = 120201
	// ErrCodeRedeemCodeAlreadyRedeemed 兑换码已兑换
	ErrCodeRedeemCodeAlreadyRedeemed = 120202
	// ErrCodeRedeemCodeCreateFailed 兑换码创建失败
	ErrCodeRedeemCodeCreateFailed = 120203
	// ErrCodeRedeemCodeRedeemFailed 兑换码兑换失败
	ErrCodeRedeemCodeRedeemFailed = 120204
	// ErrCodeRedeemCodeAssignFailed 兑换码分配失败
	ErrCodeRedeemCodeAssignFailed = 120205
)

// 奖励模块错误码 (120300-120399)
const (
	// ErrCodeRewardNotFound 奖励不存在
	ErrCodeRewardNotFound = 120301
	// ErrCodeRewardCreateFailed 奖励创建失败
	ErrCodeRewardCreateFailed = 120302
	// ErrCodeRewardUpdateFailed 奖励更新失败
	ErrCodeRewardUpdateFailed = 120303
	// ErrCodeRewardDeleteFailed 奖励删除失败
	ErrCodeRewardDeleteFailed = 120304
)

// 任务模块错误码 (120400-120499)
const (
	// ErrCodeTaskNotFound 任务不存在
	ErrCodeTaskNotFound = 120401
	// ErrCodeTaskCreateFailed 任务创建失败
	ErrCodeTaskCreateFailed = 120402
	// ErrCodeTaskUpdateFailed 任务更新失败
	ErrCodeTaskUpdateFailed = 120403
	// ErrCodeTaskDeleteFailed 任务删除失败
	ErrCodeTaskDeleteFailed = 120404
	// ErrCodeTaskTriggerFailed 任务触发失败
	ErrCodeTaskTriggerFailed = 120405
)

// 受众模块错误码 (120500-120599)
const (
	// ErrCodeAudienceNotFound 受众不存在
	ErrCodeAudienceNotFound = 120501
	// ErrCodeAudienceCreateFailed 受众创建失败
	ErrCodeAudienceCreateFailed = 120502
	// ErrCodeAudienceUpdateFailed 受众更新失败
	ErrCodeAudienceUpdateFailed = 120503
	// ErrCodeAudienceDeleteFailed 受众删除失败
	ErrCodeAudienceDeleteFailed = 120504
)

// 通知模块错误码 (120600-120699)
const (
	// ErrCodeNotificationServiceUnavailable 通知服务不可用
	ErrCodeNotificationServiceUnavailable = 120601
	// ErrCodeNotificationSendFailed 通知发送失败
	ErrCodeNotificationSendFailed = 120602
)

// 分发器模块错误码 (120700-120799)
const (
	// ErrCodeDistributorNotFound 分发器不存在
	ErrCodeDistributorNotFound = 120701
	// ErrCodeWebhookURLNotConfigured Webhook URL 未配置
	ErrCodeWebhookURLNotConfigured = 120702
	// ErrCodeWebhookRequestFailed Webhook 请求失败
	ErrCodeWebhookRequestFailed = 120703
)

