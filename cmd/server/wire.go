//go:build wireinject
// +build wireinject

package main

import (
	"marketing-service/conf"
	"marketing-service/internal/biz"
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
		biz.ProviderSet,
		// 手动添加需要特殊依赖的服务
		biz.NewAudienceMatcherService, // 需要 AudienceRepo
		biz.NewValidatorService,       // 需要 AudienceMatcherService
		// NewTaskTriggerService 需要 RocketMQ Producer，在 wire_gen.go 中需要传入 nil 或实际的 Producer
		biz.NewTaskTriggerService, // 需要所有依赖（包括 RocketMQ Producer，可为 nil）
		service.ProviderSet,
		newApp,
	))
}
