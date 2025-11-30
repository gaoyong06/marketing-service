# Marketing Service 营销服务

营销服务是一个通用的营销中台服务，核心职责是促销转化，提供活动管理、任务系统、奖励发放、兑换码等通用权益功能。

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

### 四大核心实体

1. **活动（Campaign）** - 定义营销活动的业务场景和生命周期
2. **受众（Audience）** - 定义活动的参与人群（标签、分段、列表等）
3. **任务（Task）** - 定义用户需要完成的任务（邀请、购买、分享、签到等）
4. **奖励（Reward）** - 定义奖励模板（优惠券、积分、兑换码、订阅等）

### 三大配置组件

通过 JSON 配置实现，存储在 Reward 表中：

1. **生成器（Generator）** - 奖励内容生成（兑换码、优惠券、积分）
2. **校验器（Validator）** - 奖励发放校验（时间、用户、频次、库存）
3. **发放器（Distributor）** - 奖励发放方式（自动、Webhook、邮件、短信）

### 核心业务流程

- ✅ **活动管理** - 创建、查询、更新、删除营销活动
- ✅ **任务系统** - 任务创建、事件触发、完成条件检查、自动奖励发放
- ✅ **奖励发放** - 完整的奖励发放流程（校验 → 库存预占 → 生成 → 发放）
- ✅ **兑换码管理** - 批量生成、分配、核销兑换码
- ✅ **库存管理** - 库存预占、确认、取消，防止超发
- ✅ **事件追踪** - 任务完成日志、奖励发放记录、统计分析

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
│  - campaign.go, reward.go, task.go     │
│  - task_trigger.go (事件触发服务)        │
│  - validator.go, generator.go          │
│  - distributor.go                       │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│      Data Layer (Repository)             │
│  internal/data/                          │
│  - campaign.go, reward.go, task.go     │
│  - cache.go (Redis 缓存)                 │
│  - migration.go (数据库迁移)              │
└─────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────┐
│         Database (MySQL)                │
│          Cache (Redis)                  │
└─────────────────────────────────────────┘
```

### 设计原则

- **积木式设计**: 四个核心实体（Campaign、Audience、Task、Reward）可以自由组合
- **极简主义**: 遵循"至繁归于至简"的设计哲学
- **配置化**: 轻量级逻辑通过 JSON 配置实现，避免过度设计
- **多租户**: 支持 `tenant_id` + `app_id` 双维度资源隔离

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

# 数据库会自动创建所有表和索引（包括性能优化索引）
```

### 配置服务

编辑 `configs/config.yaml`：

```yaml
server:
  http:
    addr: 0.0.0.0:8105
  grpc:
    addr: 0.0.0.0:9105

data:
  database:
    source: root:password@tcp(127.0.0.1:3306)/marketing_service?charset=utf8mb4&parseTime=True&loc=Local
  redis:
    addr: 127.0.0.1:6379
```

### 启动服务

```bash
# 生成 Proto 代码
make api

# 生成 Wire 代码
make wire

# 启动服务
make run
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

#### 活动管理 (Campaign)

- `POST /v1/campaigns` - 创建活动
- `GET /v1/campaigns/{campaign_id}` - 获取活动
- `GET /v1/campaigns` - 列出活动
- `PUT /v1/campaigns/{campaign_id}` - 更新活动
- `DELETE /v1/campaigns/{campaign_id}` - 删除活动

#### 奖励管理 (Reward)

- `POST /v1/rewards` - 创建奖励
- `GET /v1/rewards/{reward_id}` - 获取奖励
- `GET /v1/rewards` - 列出奖励
- `PUT /v1/rewards/{reward_id}` - 更新奖励
- `DELETE /v1/rewards/{reward_id}` - 删除奖励

#### 任务管理 (Task)

- `POST /v1/tasks` - 创建任务
- `GET /v1/tasks/{task_id}` - 获取任务
- `GET /v1/tasks` - 列出任务
- `PUT /v1/tasks/{task_id}` - 更新任务
- `DELETE /v1/tasks/{task_id}` - 删除任务
- `GET /v1/campaigns/{campaign_id}/tasks` - 列出活动的任务
- `POST /v1/tasks/trigger` - 触发任务事件

#### 受众管理 (Audience)

- `POST /v1/audiences` - 创建受众
- `GET /v1/audiences/{audience_id}` - 获取受众
- `GET /v1/audiences` - 列出受众
- `PUT /v1/audiences/{audience_id}` - 更新受众
- `DELETE /v1/audiences/{audience_id}` - 删除受众

#### 兑换码管理 (RedeemCode)

- `POST /v1/campaigns/{campaign_id}/redeem-codes` - 生成兑换码
- `POST /v1/redeem` - 兑换码核销
- `POST /v1/redeem-codes/{code}/assign` - 分配兑换码
- `GET /v1/redeem-codes` - 列出兑换码
- `GET /v1/redeem-codes/{code}` - 获取兑换码

#### 奖励发放 (RewardGrant)

- `GET /v1/reward-grants` - 列出奖励发放记录
- `GET /v1/reward-grants/{grant_id}` - 获取奖励发放记录
- `PUT /v1/reward-grants/{grant_id}/status` - 更新发放状态

#### 库存管理 (Inventory)

- `POST /v1/inventory/reserve` - 预占库存
- `POST /v1/inventory/{reservation_id}/confirm` - 确认库存
- `POST /v1/inventory/{reservation_id}/cancel` - 取消库存
- `GET /v1/inventory/reservations` - 列出库存预占记录

#### 任务完成日志 (TaskCompletionLog)

- `GET /v1/task-completion-logs` - 列出任务完成日志
- `GET /v1/tasks/{task_id}/completion-stats` - 获取任务完成统计

#### 活动-任务关联 (CampaignTask)

- `POST /v1/campaigns/{campaign_id}/tasks` - 将任务添加到活动
- `DELETE /v1/campaigns/{campaign_id}/tasks/{task_id}` - 从活动中移除任务
- `GET /v1/campaigns/{campaign_id}/tasks` - 列出活动的所有任务

### API 示例

#### 创建活动

```bash
curl -X POST http://localhost:8105/v1/campaigns \
  -H "Content-Type: application/json" \
  -d '{
    "tenant_id": "tenant1",
    "product_code": "app1",
    "campaign_name": "双十一促销",
    "campaign_type": "PROMOTION",
    "start_time": "2024-11-01T00:00:00Z",
    "end_time": "2024-11-11T23:59:59Z"
  }'
```

#### 触发任务事件

```bash
curl -X POST http://localhost:8105/v1/tasks/trigger \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "USER_REGISTER",
    "user_id": 123,
    "tenant_id": "tenant1",
    "product_code": "app1",
    "event_data": {
      "count": 1
    }
  }'
```

---

## 🗄️ 数据库设计

### 数据库表结构

**核心实体表（4张）**:
- `campaign` - 活动表
- `audience` - 受众表
- `task` - 任务表
- `reward` - 奖励表

**关系表（1张）**:
- `campaign_task` - 活动-任务关联表

**业务数据表（4张）**:
- `reward_grant` - 奖励发放表
- `redeem_code` - 兑换码表
- `task_completion_log` - 任务完成日志表
- `inventory_reservation` - 库存预占表

### 数据库初始化

```bash
# 创建数据库和表结构（包含性能优化索引）
mysql -u root -p < docs/sql/marketing_service.sql
```

### 性能优化

数据库已包含以下性能优化索引：

- **复合索引**: 租户+应用+状态（用于列表查询）
- **时间范围索引**: 用于查询活跃任务和活动
- **用户索引**: 用于用户维度的查询
- **统计索引**: 用于各种统计查询

详细索引定义请参考 `docs/sql/marketing_service.sql`。

---

## 📁 项目结构

```
marketing-service/
├── api/                          # API 定义
│   ├── base/                     # 基础类型（错误、分页）
│   └── marketing_service/v1/     # 营销服务 API
│       └── marketing.proto      # API 定义文件
├── cmd/                          # 服务入口
│   └── marketing-service/        # 主程序
│       ├── main.go              # 启动入口
│       ├── wire.go              # Wire 配置
│       └── wire_gen.go          # Wire 生成代码
├── configs/                      # 配置文件
│   └── config.yaml              # 服务配置
├── docs/                         # 文档
│   ├── product_design.md        # 产品设计文档
│   ├── logic_design.md          # 业务逻辑设计
│   ├── IMPLEMENTATION_PLAN.md   # 实施计划
│   ├── IMPLEMENTATION_PROGRESS.md # 实施进度
│   ├── MISSING_FEATURES.md      # 功能清单
│   └── sql/                     # SQL 脚本
│       └── marketing_service.sql # 数据库脚本（包含索引优化）
├── internal/                     # 内部代码
│   ├── biz/                     # 业务逻辑层
│   │   ├── campaign.go          # 活动业务逻辑
│   │   ├── reward.go            # 奖励业务逻辑
│   │   ├── task.go              # 任务业务逻辑
│   │   ├── task_trigger.go      # 任务触发服务
│   │   ├── validator.go         # 校验器
│   │   ├── generator.go         # 生成器
│   │   ├── distributor.go       # 发放器
│   │   └── ...
│   ├── data/                    # 数据访问层
│   │   ├── campaign.go          # 活动 Repository
│   │   ├── reward.go            # 奖励 Repository
│   │   ├── task.go              # 任务 Repository
│   │   ├── cache.go             # 缓存服务
│   │   ├── migration.go         # 数据库迁移
│   │   └── model/               # 数据模型
│   ├── service/                 # 服务层
│   │   └── marketing.go        # 营销服务实现
│   ├── server/                  # 服务器配置
│   │   ├── http.go              # HTTP 服务器
│   │   └── grpc.go              # gRPC 服务器
│   └── metrics/                 # 监控指标
│       └── metrics.go           # Prometheus 指标定义
├── test/                         # 测试
│   └── api/                     # API 测试
│       └── api-test-config.yaml # api-tester 测试配置（20个场景）
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

# 启动服务
make run

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
3. **实现业务逻辑** (`internal/biz/`)
4. **实现数据层** (`internal/data/`)
5. **实现服务层** (`internal/service/`)
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

**测试场景**（20个）:
- 基础功能测试（健康检查、Metrics）
- Campaign/Reward/Task/Audience CRUD 测试
- 任务触发和奖励发放流程测试
- 兑换码功能测试
- 库存管理测试
- 任务完成日志查询测试
- 完整业务流程测试

---

## 📊 监控和日志

### Prometheus 指标

服务暴露 Prometheus 指标端点：`GET /metrics`

**业务指标**:
- `marketing_campaign_created_total` - 活动创建数量
- `marketing_task_created_total` - 任务创建数量
- `marketing_task_triggered_total` - 任务触发数量
- `marketing_task_completed_total` - 任务完成数量
- `marketing_reward_created_total` - 奖励创建数量
- `marketing_reward_granted_total` - 奖励发放数量
- `marketing_redeem_code_generated_total` - 兑换码生成数量
- `marketing_redeem_code_redeemed_total` - 兑换码兑换数量
- `marketing_inventory_reserved_total` - 库存预占数量
- `marketing_inventory_confirmed_total` - 库存确认数量
- `marketing_inventory_cancelled_total` - 库存取消数量

**性能指标**:
- `marketing_task_trigger_duration_seconds` - 任务触发耗时
- `marketing_reward_generation_duration_seconds` - 奖励生成耗时
- `marketing_reward_distribution_duration_seconds` - 奖励发放耗时

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
export MARKETING_DB_SOURCE="root:password@tcp(127.0.0.1:3306)/marketing_service"
export MARKETING_REDIS_ADDR="127.0.0.1:6379"
export MARKETING_HTTP_PORT="8105"
export MARKETING_GRPC_PORT="9105"
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
- [实施计划](docs/IMPLEMENTATION_PLAN.md) - 代码实施计划
- [实施进度](docs/IMPLEMENTATION_PROGRESS.md) - 实施进度跟踪
- [功能清单](docs/MISSING_FEATURES.md) - 功能完成度清单

---

## 🎯 核心特性

### 1. 积木式设计

四个核心实体（Campaign、Audience、Task、Reward）可以自由组合，构建复杂的营销场景。

### 2. 配置化组件

Generator、Validator、Distributor 通过 JSON 配置实现，无需独立表，灵活可扩展。

### 3. 完整的奖励发放流程

- **校验阶段**: 时间、用户、频次、库存校验
- **生成阶段**: 兑换码、优惠券、积分生成
- **发放阶段**: 自动、Webhook、邮件、短信发放

### 4. 性能优化

- **缓存层**: Campaign、Reward、Task 缓存（Redis）
- **批量操作**: 批量生成兑换码、批量发放奖励
- **数据库优化**: 复合索引、分页查询优化

### 5. 多租户支持

所有 API 和数据库操作都支持 `tenant_id` + `app_id` 双维度资源隔离。

---

## 📈 项目状态

### ✅ 已完成功能

- ✅ 所有核心实体 CRUD API
- ✅ 任务触发和奖励发放完整流程
- ✅ 兑换码管理
- ✅ 库存管理
- ✅ 活动-任务关联管理
- ✅ 缓存层（Redis）
- ✅ Prometheus 监控指标
- ✅ API 集成测试（20个测试场景）

### 📊 功能完成度

- **P0（核心功能）**: 100% ✅
- **P1（扩展功能）**: 100% ✅
- **P2（高级功能）**: 100% ✅

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
**版本**: v1.0.0
