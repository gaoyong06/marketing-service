# Marketing Service - go-pkg 集成说明

## 已集成的 go-pkg 组件

### 1. ✅ Error 处理 (`go-pkg/errors`)
- **位置**: `internal/service/marketing.go`, `internal/data/*.go`
- **使用**: `pkgErrors.NewBizError(pkgErrors.ErrCodeNotFound, "zh-CN")`
- **说明**: 统一错误码和错误处理

### 2. ✅ Logger (`go-pkg/logger`)
- **位置**: `cmd/server/main.go`
- **使用**: `logger.NewLogger(logCfg)`
- **说明**: 结构化日志，支持轮转

### 3. ✅ Health 检查 (`go-pkg/health`)
- **位置**: `internal/server/http.go`
- **使用**: `health.NewResponse("marketing-service")`
- **说明**: 统一健康检查响应格式

### 4. ✅ Response 中间件 (`go-pkg/middleware/response`)
- **位置**: `internal/server/http.go`
- **使用**: 
  - `response.NewResponseEncoder(errorHandler, responseConfig)`
  - `response.NewErrorEncoder(errorHandler)`
- **说明**: 统一 API 响应格式

### 5. ✅ i18n 中间件 (`go-pkg/middleware/i18n`)
- **位置**: `internal/server/http.go`
- **使用**: `i18n.Middleware()`
- **说明**: 国际化支持

## 已删除的 base 目录

- ✅ 删除了 `api/base/error.proto`
- ✅ 删除了 `api/base/pagination.proto`
- ✅ 删除了生成的 `.pb.go` 文件
- ✅ 所有错误处理统一使用 `go-pkg/errors`

## 参考实现

参考 `passport-service` 的实现方式：
- 不使用 base 目录
- 统一使用 go-pkg 组件
- 响应格式统一
