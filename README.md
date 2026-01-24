# Marketing Service 营销服务（极简重构版）

营销服务是一个极简的优惠券管理服务，专注于支付场景的优惠券管理和使用统计。遵循"少即是多"的设计理念，仅保留核心优惠券功能。

## 📋 目录

- [核心功能](#核心功能)
- [技术架构](#技术架构)
- [快速开始](#快速开始)
- [API 文档](#api-文档)
- [数据库设计](#数据库设计)
- [项目结构](#项目结构)
- [开发指南](#开发指南)
- [测试](#测试)
- [部署](#部署)
- [监控和日志](#监控和日志)

---

## 🎯 核心功能

### 优惠券管理

- ✅ **优惠券 CRUD** - 创建、查询、更新、删除优惠券
- ✅ **优惠券验证** - 供 Payment Service 调用，验证优惠券有效性
- ✅ **优惠券使用** - 记录优惠券使用情况，更新使用次数
- ✅ **使用记录** - 记录每次优惠券使用的详细信息
- ✅ **统计分析** - 优惠券使用统计、转化率分析、汇总统计

### 设计理念

- **极简主义** - 只保留核心优惠券功能，移除复杂营销活动系统
- **专注支付场景** - 专注于支付场景的优惠券管理和使用
- **克制设计** - 避免过度设计，保持简单易用

---

## 🏗️ 技术架构

### 技术栈

- **框架**: Kratos (Go 微服务框架)
- **ORM**: GORM v2
- **数据库**: MySQL 8.0+
- **缓存**: Redis
- **依赖注入**: Wire
- **API 定义**: Protocol Buffers (gRPC + HTTP)
- **监控**: Prometheus
- **测试**: api-tester

### 架构设计

```
┌─────────────────────────────────────────┐
│         API Layer (gRPC/HTTP)          │
│     internal/service/marketing.go      │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│      Business Layer (UseCase)           │
│  internal/biz/                          │
│  - coupon.go (优惠券业务逻辑)            │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│      Data Layer (Repository)             │
│  internal/data/                          │
│  - coupon.go (优惠券 Repository)         │
│  - model/coupon.go (数据模型)             │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Database (MySQL)                │
│          Cache (Redis)                  │
└─────────────────────────────────────────┘
```

### 设计原则

- **极简主义**: 遵循"至繁归于至简"的设计哲学
- **业务导向**: 专注于支付场景的优惠券管理和使用
- **克制设计**: 避免过度设计，保持简单易用

---

## 🚀 快速开始

### 环境要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+
- Make

### 安装依赖

```bash
# 克隆项目
git clone <repository-url>
cd marketing-service

# 安装 Go 依赖
go mod download
```

### 配置数据库

```bash
# 初始化数据库
mysql -u root -p < docs/sql/marketing_service.sql

# 数据库会自动创建所有表和索引
```

### 配置服务

编辑 `configs/config_debug.yaml`（开发环境）或 `configs/config_release.yaml`（生产环境）：

```yaml
server:
  http:
    addr: 0.0.0.0:8105
  grpc:
    addr: 0.0.0.0:9105

data:
  database:
    source: root:@tcp(127.0.0.1:3306)/marketing_service?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
```

### 启动服务

```bash
# 生成 Proto 代码
make api

# 生成 Wire 代码
make wire

# 启动服务（开发模式，默认使用 config_debug.yaml）
make run

# 或启动生产模式（使用 config_release.yaml）
./bin/server -mode release
```

### 验证服务

```bash
# 健康检查
curl http://localhost:8105/health

# Prometheus 指标
curl http://localhost:8105/metrics
```

---

## 📚 API 文档

### 基础端点

- **健康检查**: `GET /health`
- **Prometheus 指标**: `GET /metrics`

### 核心 API

#### 优惠券管理 (Coupon)

- `POST /v1/coupons` - 创建优惠券
- `GET /v1/coupons/{couponCode}` - 获取优惠券
- `GET /v1/coupons` - 列出优惠券（支持按 appId 和 status 筛选）
- `PUT /v1/coupons/{couponCode}` - 更新优惠券
- `DELETE /v1/coupons/{couponCode}` - 删除优惠券

#### 优惠券验证和使用（供 Payment Service 调用）

- `POST /v1/coupons/validate` - 验证优惠券有效性
- `POST /v1/coupons/use` - 使用优惠券（记录使用情况）

#### 统计分析

- `GET /v1/coupons/{couponCode}/stats` - 获取单个优惠券统计
- `GET /v1/coupons/{couponCode}/usages` - 列出优惠券使用记录
- `GET /v1/coupons/summary-stats` - 获取所有优惠券汇总统计（按 appId 筛选）

### API 示例

#### 创建优惠券

```bash
curl -X POST http://localhost:8105/v1/coupons \
  -H "Content-Type: application/json" \
  -d '{
    "couponCode": "WELCOME10",
    "discountType": "percent",
    "discountValue": 10,
    "validFrom": 1733011200,
    "validUntil": 1735689599,
    "maxUses": 1000,
    "minAmount": 10000
  }'
```

#### 验证优惠券

```bash
curl -X POST http://localhost:8105/v1/coupons/validate \
  -H "Content-Type: application/json" \
  -d '{
    "couponCode": "WELCOME10",
    "amount": 20000
  }'
```

#### 使用优惠券

```bash
curl -X POST http://localhost:8105/v1/coupons/use \
  -H "Content-Type: application/json" \
  -d '{
    "couponCode": "WELCOME10",
    "appId": "app123",
    "userId": "user123",
    "paymentOrderId": "order123",
    "paymentId": "payment123",
    "originalAmount": 20000,
    "discountAmount": 2000,
    "finalAmount": 18000
  }'
```

#### 获取优惠券统计

```bash
curl http://localhost:8105/v1/coupons/WELCOME10/stats
```

---

## 🗄️ 数据库设计

### 数据库表结构

**核心表（2张）**:
- `coupon` - 优惠券表
- `coupon_usage` - 优惠券使用记录表

### 数据库初始化

```bash
# 创建数据库和表结构（包含性能优化索引）
mysql -u root -p < docs/sql/marketing_service.sql
```

### 性能优化

数据库已包含以下性能优化索引：

- **唯一索引**: `coupon_code`（全局唯一，严格限制重复，即使软删除的记录也不允许 code 重复）
- **应用索引**: `app_id`（用于按应用查询）
- **状态索引**: `status`（用于状态筛选）
- **时间范围索引**: `valid_from`, `valid_until`（用于有效期查询）
- **使用记录索引**: `coupon_code`, `app_id`, `user_id`, `payment_order_id`, `payment_id`, `used_at`（用于各种查询场景）

详细索引定义请参考 `docs/sql/marketing_service.sql`。

---

## 📁 项目结构

```
marketing-service/
├── api/                          # API 定义
│   └── marketing_service/v1/     # 营销服务 API
│       ├── marketing.proto       # API 定义文件
│       ├── marketing.pb.go       # 生成的 Go 代码
│       ├── marketing_grpc.pb.go  # gRPC 代码
│       └── marketing_http.pb.go  # HTTP 代码
├── cmd/                          # 服务入口
│   └── server/                   # 主程序
│       ├── main.go              # 启动入口
│       ├── wire.go              # Wire 配置
│       └── wire_gen.go          # Wire 生成代码
├── configs/                      # 配置文件
│   ├── config_debug.yaml        # 开发环境配置
│   └── config_release.yaml      # 生产环境配置
├── docs/                         # 文档
│   ├── product_design.md        # 产品设计文档
│   ├── logic_design.md          # 业务逻辑设计
│   └── sql/                     # SQL 脚本
│       └── marketing_service.sql # 数据库脚本（包含索引优化）
├── internal/                     # 内部代码
│   ├── biz/                     # 业务逻辑层
│   │   ├── coupon.go           # 优惠券业务逻辑
│   │   └── utils.go            # 工具函数
│   ├── data/                    # 数据访问层
│   │   ├── coupon.go           # 优惠券 Repository
│   │   ├── data.go             # 数据层初始化
│   │   └── model/              # 数据模型
│   │       └── coupon.go       # 优惠券模型
│   ├── service/                 # 服务层
│   │   └── marketing.go        # 营销服务实现
│   ├── server/                  # 服务器配置
│   │   ├── http.go              # HTTP 服务器
│   │   └── grpc.go              # gRPC 服务器
│   ├── metrics/                 # 监控指标
│   │   └── metrics.go           # Prometheus 指标定义
│   ├── errors/                  # 错误定义
│   │   └── code.go              # 错误码
│   ├── constants/               # 常量定义
│   │   └── constants.go         # 业务常量
│   └── conf/                     # 配置定义
│       ├── conf.proto           # 配置 Proto
│       └── conf.pb.go           # 配置 Go 代码
├── i18n/                         # 国际化
│   ├── zh-CN/                   # 中文错误信息
│   │   └── errors.json
│   └── en-US/                    # 英文错误信息
│       └── errors.json
├── test/                         # 测试
│   └── api/                     # API 测试
│       └── api-test-config.yaml # api-tester 测试配置
├── Makefile                      # 构建脚本
├── go.mod                        # Go 模块定义
└── README.md                     # 项目说明（本文件）
```

---

## 🛠️ 开发指南

### 常用命令

```bash
# 生成 Proto 代码
make api

# 生成 Wire 代码
make wire

# 运行测试
make test

# 启动服务（开发模式）
make run

# 启动服务（生产模式）
./bin/server -mode release

# 构建二进制文件
make build
```

### 代码生成

```bash
# 生成 Proto 代码（gRPC + HTTP）
make api

# 生成 Wire 依赖注入代码
make wire
```

### 开发流程

1. **修改 Proto 定义** (`api/marketing_service/v1/marketing.proto`)
2. **生成代码**: `make api`
3. **实现业务逻辑** (`internal/biz/coupon.go`)
4. **实现数据层** (`internal/data/coupon.go`)
5. **实现服务层** (`internal/service/marketing.go`)
6. **更新 Wire 配置**: `make wire`
7. **运行测试**: `make test`

---

## 🧪 测试

### 单元测试

```bash
# 运行所有单元测试
go test ./internal/... -v

# 运行特定包的测试
go test ./internal/biz/... -v
go test ./internal/data/... -v
```

### 集成测试

项目使用 `api-tester` 进行 API 集成测试：

```bash
# 运行 API 测试
make test

# 测试配置文件
test/api/api-test-config.yaml
```

**测试场景**:
- 基础功能测试（健康检查、Metrics）
- 优惠券 CRUD 测试
- 优惠券验证和使用流程测试
- 统计分析功能测试
- 完整业务流程测试

---

## 📊 监控和日志

### Prometheus 指标

服务暴露 Prometheus 指标端点：`GET /metrics`

**业务指标**:
- `marketing_coupon_created_total` - 优惠券创建数量
- `marketing_coupon_validated_total` - 优惠券验证数量
- `marketing_coupon_used_total` - 优惠券使用数量

**性能指标**:
- `marketing_coupon_validate_duration_seconds` - 优惠券验证耗时
- `marketing_coupon_use_duration_seconds` - 优惠券使用耗时

### 健康检查

```bash
curl http://localhost:8105/health
```

响应：
```json
{
  "status": "UP",
  "service": "marketing-service"
}
```

### 日志

服务使用 Kratos log 框架，支持结构化日志。日志级别可通过配置调整。

---

## 🚢 部署

### 环境变量

可通过环境变量覆盖配置：

```bash
export MARKETING_DB_SOURCE="root:@tcp(127.0.0.1:3306)/marketing_service"
export MARKETING_REDIS_ADDR="127.0.0.1:6379"
export MARKETING_HTTP_PORT="8105"
export MARKETING_GRPC_PORT="9105"
```

### 运行模式

服务支持两种运行模式：

```bash
# 开发模式（默认，使用 config_debug.yaml）
./bin/server

# 或显式指定
./bin/server -mode debug

# 生产模式（使用 config_release.yaml）
./bin/server -mode release

# 兼容旧方式（仍可使用，但不推荐）
./bin/server -conf ../../configs/config_release.yaml
```

### Docker 部署

```bash
# 构建镜像
docker build -t marketing-service:latest .

# 运行容器
docker run -d \
  -p 8105:8105 \
  -p 9105:9105 \
  -e MARKETING_DB_SOURCE="..." \
  -e MARKETING_REDIS_ADDR="..." \
  marketing-service:latest
```

### 生产环境建议

1. **数据库**: 使用主从复制，读写分离
2. **缓存**: Redis 集群模式
3. **监控**: 集成 Prometheus + Grafana
4. **日志**: 集成 ELK 或类似日志系统
5. **限流**: 在网关层实现 API 限流
6. **安全**: 使用 TLS/SSL 加密通信

---

## 📖 设计文档

- [产品设计文档](docs/product_design.md) - 产品设计理念和功能说明
- [业务逻辑设计](docs/logic_design.md) - 业务逻辑和流程设计

---

## 🎯 核心特性

### 1. 极简设计

只保留核心优惠券功能，专注于支付场景的优惠券管理和使用。

### 2. 完整的优惠券生命周期

- **创建**: 支持百分比折扣和固定金额折扣
- **验证**: 供 Payment Service 调用，验证优惠券有效性
- **使用**: 记录使用情况，更新使用次数
- **统计**: 使用统计、转化率分析、汇总统计

### 3. 性能优化

- **缓存层**: 优惠券信息缓存（Redis）
- **数据库优化**: 复合索引、分页查询优化
- **事务支持**: 确保使用记录和计数更新的原子性

### 4. 多应用支持

所有 API 和数据库操作都支持 `app_id` 维度资源隔离。

---

## 📈 项目状态

### ✅ 已完成功能

- ✅ 优惠券 CRUD API
- ✅ 优惠券验证和使用完整流程
- ✅ 使用记录查询
- ✅ 统计分析功能
- ✅ Prometheus 监控指标
- ✅ API 集成测试

### 📊 功能完成度

- **P0（核心功能）**: 100% ✅
- **P1（扩展功能）**: 100% ✅

**总体完成度**: 100% ✅

---

## 🤝 贡献指南

1. Fork 项目
2. 创建功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

[添加许可证信息]

---

## 📞 联系方式

[添加联系方式]

---

## 🙏 致谢

感谢所有为这个项目做出贡献的开发者！

---

**最后更新**: 2024-12-XX
**版本**: v2.0.0 (极简重构版)
