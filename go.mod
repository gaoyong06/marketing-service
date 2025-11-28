module github.com/gaoyong06/middleground/marketing-service

go 1.23

toolchain go1.23.8

require (
	github.com/gaoyong06/middleground/proto-repo v0.0.0-00010101000000-000000000000
	github.com/go-kratos/kratos/v2 v2.7.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	github.com/google/wire v0.5.0
	google.golang.org/grpc v1.72.0
	google.golang.org/protobuf v1.36.5
	gorm.io/driver/mysql v1.5.2
	gorm.io/gorm v1.25.5
)

// 使用 replace 指令来处理本地模块
replace (
	github.com/gaoyong06/middleground => ../
	github.com/gaoyong06/middleground/marketing-service => ./

	// 本地开发时使用本地路径
	github.com/gaoyong06/middleground/proto-repo => ../proto-repo
	github.com/gaoyong06/middleground/tenant-service => ../tenant-service
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/envoyproxy/protoc-gen-validate v1.2.1 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-kratos/aegis v0.2.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/form/v4 v4.2.1 // indirect
	github.com/go-sql-driver/mysql v1.7.0 // indirect
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/imdario/mergo v0.3.16 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	go.opentelemetry.io/auto/sdk v1.1.0 // indirect
	go.opentelemetry.io/otel v1.34.0 // indirect
	go.opentelemetry.io/otel/metric v1.34.0 // indirect
	go.opentelemetry.io/otel/trace v1.34.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sync v0.11.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250218202821-56aae31c358a // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250218202821-56aae31c358a // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
