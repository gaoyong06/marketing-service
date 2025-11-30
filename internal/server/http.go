package server

import (
	"marketing-service/conf"

	"github.com/go-kratos/kratos/v2/log"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/service"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(s *conf.Server, marketing *service.MarketingService, logger log.Logger) *kratoshttp.Server {
	var opts []kratoshttp.ServerOption
	if s != nil && s.Http != nil {
		if s.Http.Addr != "" {
			opts = append(opts, kratoshttp.Address(s.Http.Addr))
		}
		if s.Http.Timeout != nil {
			opts = append(opts, kratoshttp.Timeout(s.Http.Timeout.AsDuration()))
		}
	}

	srv := kratoshttp.NewServer(opts...)
	v1.RegisterMarketingHTTPServer(srv, marketing)

	// 注册 Prometheus metrics 端点
	srv.Route("/").GET("/metrics", func(ctx kratoshttp.Context) error {
		promhttp.Handler().ServeHTTP(ctx.Response(), ctx.Request())
		return nil
	})

	// 注册健康检查端点
	srv.Route("/").GET("/health", func(ctx kratoshttp.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status":  "UP",
			"service": "marketing-service",
		})
	})

	return srv
}
