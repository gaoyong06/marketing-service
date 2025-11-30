package server

import (
	"marketing-service/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/service"
)

// NewGRPCServer 创建 gRPC 服务器
func NewGRPCServer(s *conf.Server, marketing *service.MarketingService, logger log.Logger) *grpc.Server {
	var opts []grpc.ServerOption
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

