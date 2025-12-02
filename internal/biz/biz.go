package biz

import "github.com/google/wire"

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewCampaignUseCase,
	NewRewardUseCase,
	NewRewardGrantUseCase,
	NewTaskUseCase,
	NewAudienceUseCase,
	NewRedeemCodeUseCase,
	NewInventoryReservationUseCase,
	NewTaskCompletionLogUseCase,
	NewCampaignTaskUseCase,
	NewGeneratorService,
	NewDistributorService,
	NewNotificationService,
	// 注意：以下服务需要特殊依赖，在 wire.go 中手动构建
	// NewAudienceMatcherService, // 需要 AudienceRepo
	// NewValidatorService,        // 需要 AudienceMatcherService
	// NewTaskTriggerService,     // 需要所有依赖（包括 RocketMQ Producer，可为 nil）
)
