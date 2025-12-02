package server

import (
	"marketing-service/internal/conf"

	"github.com/gaoyong06/go-pkg/middleware/i18n"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/service"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(s *conf.Server, marketing *service.MarketingService, logger log.Logger) *grpc.Server {
	var opts []grpc.ServerOption

	// 添加中间件：recovery、i18n
	opts = append(opts, grpc.Middleware(
		recovery.Recovery(),
		i18n.Middleware(), // 国际化中间件
	))

	if s != nil && s.Grpc != nil {
		if s.Grpc.Addr != "" {
			opts = append(opts, grpc.Address(s.Grpc.Addr))
		}
		if s.Grpc.Timeout != nil {
			opts = append(opts, grpc.Timeout(s.Grpc.Timeout.AsDuration()))
		}
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterMarketingServer(srv, marketing)
	return srv
}
