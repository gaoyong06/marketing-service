package main

import (
	"flag"
	"fmt"
	"os"

	"marketing-service/internal/conf"
	_ "marketing-service/internal/errors" // 初始化错误管理器

	"github.com/gaoyong06/go-pkg/logger"
	pkgutils "github.com/gaoyong06/go-pkg/utils"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string = "marketing-service"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
	// runMode is the run mode (debug, release).
	runMode string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "", "config path, eg: -conf config.yaml (deprecated, use -mode instead)")
	flag.StringVar(&runMode, "mode", "debug", "Run mode (debug, release)")
}

func newApp(logger log.Logger, hs *http.Server, gs *grpc.Server) *kratos.App {
	return kratos.New(
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			hs,
			gs,
		),
	)
}

func main() {
	flag.Parse()

	// 根据 mode 自动选择配置文件
	configPath := flagconf
	if configPath == "" {
		// 使用 go-pkg/utils 中的通用配置文件路径解析函数
		// 支持从不同目录运行（项目根目录、cmd/server 目录等）
		configPath = pkgutils.FindConfigFileWithMode(runMode, []string{
			"configs",       // 从项目根目录运行
			"../../configs", // 从 cmd/server 目录运行
			"../configs",    // 从 cmd 目录运行
		})
	}

	c := config.New(
		config.WithSource(
			file.NewSource(configPath),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	// 验证配置
	if err := bc.Validate(); err != nil {
		panic(fmt.Sprintf("config validation failed: %v", err))
	}

	logCfg := &logger.Config{
		Level:    "info",
		Format:   "json",
		Output:   "stdout",
		FilePath: "",
	}
	appLogger := logger.NewLogger(logCfg)
	appLogger = log.With(appLogger,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
	)

	app, cleanup, err := wireApp(bc.Server, bc.Data, bc.Client, appLogger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
