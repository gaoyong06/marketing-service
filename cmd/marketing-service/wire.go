//go:build wireinject
// +build wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"marketing-service/internal/biz"
	"marketing-service/internal/conf"
	"marketing-service/internal/data"
	"marketing-service/internal/server"
	"marketing-service/internal/service"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Client, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, data.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}
