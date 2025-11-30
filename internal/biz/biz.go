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
	NewValidatorService,
	NewGeneratorService,
	NewDistributorService,
	NewTaskTriggerService,
)
