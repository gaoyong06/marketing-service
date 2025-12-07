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
// 极简重构：仅保留优惠券功能，移除复杂营销活动系统
func wireApp(*conf.Server, *conf.Data, *conf.Client, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}
