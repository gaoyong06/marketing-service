package server

import (
	"marketing-service/conf"

	"github.com/gaoyong06/go-pkg/health"
	"github.com/gaoyong06/go-pkg/middleware/i18n"
	"github.com/gaoyong06/go-pkg/middleware/response"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/validate"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	v1 "marketing-service/api/marketing_service/v1"
	"marketing-service/internal/service"
)

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(s *conf.Server, marketing *service.MarketingService, logger log.Logger) *kratoshttp.Server {
	// 响应中间件配置
	responseConfig := &response.Config{
		EnableUnifiedResponse: true,
		IncludeDetailedError:  true, // 开发环境可以为 true
		IncludeHost:           true,
		IncludeTraceId:        true,
	}

	// 使用默认错误处理器（已支持 Kratos errors 的 HTTP 状态码映射）
	errorHandler := response.NewDefaultErrorHandler()

	var opts []kratoshttp.ServerOption

	// 添加中间件：recovery、validate、i18n
	opts = append(opts, kratoshttp.Middleware(
		recovery.Recovery(),
		validate.Validator(), // 自动验证 proto validate 规则
		i18n.Middleware(),    // 国际化中间件
	))

	// 使用自定义响应编码器统一响应格式
	opts = append(opts, kratoshttp.ResponseEncoder(response.NewResponseEncoder(errorHandler, responseConfig)))
	// 使用支持 gRPC status 的错误编码器
	opts = append(opts, kratoshttp.ErrorEncoder(response.NewErrorEncoder(errorHandler)))

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

	// 注册健康检查端点（使用 go-pkg/health）
	srv.Route("/").GET("/health", func(ctx kratoshttp.Context) error {
		return ctx.Result(200, health.NewResponse("marketing-service"))
	})

	return srv
}
