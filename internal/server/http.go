package server

import (
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
	pb "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/conf"
	"marketing-service/internal/service"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, marketingSvc *service.MarketingService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil && c.Http.Timeout.AsDuration() > 0 {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)

	// 注册服务器
	pb.RegisterMarketingHTTPServer(srv, marketingSvc)

	// 添加健康检查路由
	router := srv.Route("/")
	router.GET("/health", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]string{
			"status":  "ok",
			"service": "marketing-service",
			"version": "v1",
		})
	})

	// 添加指标路由
	router.GET("/metrics", func(ctx http.Context) error {
		return ctx.JSON(200, map[string]interface{}{
			"status": "ok",
			"uptime": "up",
		})
	})
	return srv
}
