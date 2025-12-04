//go:build wireinject
// +build wireinject

package main

import (
	"marketing-service/internal/biz"
	"marketing-service/internal/conf"
	"marketing-service/internal/data"
	"marketing-service/internal/server"
	"marketing-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Client, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		data.NewRocketMQTopic, // 提供 RocketMQ Topic
		biz.ProviderSet,
		// 手动添加需要特殊依赖的服务
		biz.NewAudienceMatcherService, // 需要 AudienceRepo
		biz.NewValidatorService,       // 需要 AudienceMatcherService
		// NewTaskTriggerService 需要 RocketMQ Producer 和 Topic
		biz.NewTaskTriggerService, // 需要所有依赖（包括 RocketMQ Producer 和 Topic）
		service.ProviderSet,
		newApp,
	))
}
