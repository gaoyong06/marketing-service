# Marketing Service 代码实施计划

## 当前状态
✅ **所有计划功能已完成！**

### 完成情况总结
- ✅ 已完成 GORM 模型定义（9个模型）
- ✅ 已完成所有 Repository 实现（13个文件）
- ✅ 已完成所有 UseCase 实现（16个文件）
- ✅ 已完成所有 API 实现（统一在 marketing.proto 和 marketing.go 中）
- ✅ 已完成服务启动配置（HTTP + gRPC）
- ✅ 已完成依赖注入（Wire）
- ✅ 已完成单元测试和集成测试
- ✅ 已完成性能优化（缓存、批量操作、数据库查询优化）
- ✅ 已完成监控和日志（Prometheus 指标）

## 实施步骤

由于完整实现需要大量代码，我建议采用**渐进式实现**策略：

### 阶段 1：核心基础设施（优先级：P0）

#### 1.1 数据层初始化
- [x] ✅ `internal/data/data.go` - Data Provider（数据库连接、Redis连接）
- [x] ✅ `internal/data/migration.go` - 数据库迁移工具

#### 1.2 核心 Repository（Campaign + Reward）
- [x] ✅ `internal/data/campaign.go` - Campaign Repository
- [x] ✅ `internal/data/reward.go` - Reward Repository  
- [x] ✅ `internal/data/reward_grant.go` - RewardGrant Repository

### 阶段 2：业务逻辑层（优先级：P0）

#### 2.1 核心业务逻辑
- [x] ✅ `internal/biz/campaign.go` - Campaign UseCase
- [x] ✅ `internal/biz/reward.go` - Reward UseCase

### 阶段 3：API 层（优先级：P0）

#### 3.1 Proto 定义
- [x] ✅ `api/marketing_service/v1/marketing.proto` - 统一 API 定义（包含 Campaign、Reward、Task、Audience 等所有 API）
  - ✅ Campaign API
  - ✅ Reward API
  - ✅ Task API
  - ✅ Audience API
  - ✅ RewardGrant API
  - ✅ 其他所有 API

#### 3.2 Service 实现
- [x] ✅ `internal/service/marketing.go` - 统一 Service 实现（包含所有业务逻辑）
  - ✅ Campaign Service
  - ✅ Reward Service
  - ✅ Task Service
  - ✅ Audience Service
  - ✅ RewardGrant Service
  - ✅ 其他所有 Service

### 阶段 4：服务启动（优先级：P0）

#### 4.1 依赖注入
- [x] ✅ `cmd/server/wire.go` - Wire 配置
- [x] ✅ `cmd/server/wire_gen.go` - Wire 生成代码

#### 4.2 服务器配置
- [x] ✅ `internal/server/http.go` - HTTP Server（包含 /metrics 和 /health 端点）
- [x] ✅ `internal/server/grpc.go` - gRPC Server

### 阶段 5：扩展功能（优先级：P1）

- [x] ✅ Task 相关实现
  - ✅ `internal/data/task.go` - Task Repository
  - ✅ `internal/biz/task.go` - Task UseCase
  - ✅ Task API（在 marketing.proto 中）
- [x] ✅ Audience 相关实现
  - ✅ `internal/data/audience.go` - Audience Repository
  - ✅ `internal/biz/audience.go` - Audience UseCase
  - ✅ Audience API（在 marketing.proto 中）
- [x] ✅ RedeemCode 相关实现
  - ✅ `internal/data/redeem_code.go` - RedeemCode Repository
  - ✅ `internal/biz/redeem_code.go` - RedeemCode UseCase
  - ✅ RedeemCode API（在 marketing.proto 中）

### 阶段 6：高级功能（优先级：P2）

- [x] ✅ 库存管理（InventoryReservation）
  - ✅ `internal/data/inventory_reservation.go` - InventoryReservation Repository
  - ✅ `internal/biz/inventory_reservation.go` - InventoryReservation UseCase
  - ✅ InventoryReservation API（在 marketing.proto 中）
- [x] ✅ 事件驱动（TaskCompletionLog）
  - ✅ `internal/data/task_completion_log.go` - TaskCompletionLog Repository
  - ✅ `internal/biz/task_completion_log.go` - TaskCompletionLog UseCase
  - ✅ TaskCompletionLog API（在 marketing.proto 中）
- [x] ✅ 单元测试
  - ✅ 核心业务逻辑单元测试（campaign_test.go, task_trigger_test.go 等）
- [x] ✅ 集成测试
  - ✅ `test/api/api-test-config.yaml` - api-tester 测试配置（19个测试场景）
  - ✅ 统一使用 api-tester 进行 API 集成测试

---

## 推荐实施方案

考虑到代码量较大，我建议：

### 方案 A：最小可用版本（MVP）
**目标**：快速搭建可运行的服务骨架

**实现内容**：
1. ✅ 数据层初始化（`data.go`）
2. ✅ Campaign CRUD Repository
3. ✅ Campaign CRUD API（Proto + Service）
4. ✅ 服务启动配置

**时间估算**：约 2-3 小时
**优势**：可以快速验证架构设计，立即可运行测试

### 方案 B：完整实现
**目标**：实现所有核心功能

**实现内容**：
- 所有 Repository
- 所有 Business Logic
- 所有 API
- 完整的单元测试

**时间估算**：约 1-2 天
**优势**：一次性完成所有功能

---

## 我的建议

**采用方案 A（MVP）+ 迭代**：

1. **第一轮**：实现 Campaign 的完整链路（Data → Biz → Service）
2. **第二轮**：实现 Reward 的完整链路
3. **第三轮**：实现 RewardGrant（奖励发放）的核心逻辑
4. **后续**：根据需要扩展其他功能

这样的好处是：
- ✅ 每一轮都能产出可运行的代码
- ✅ 可以及时发现设计问题
- ✅ 便于逐步测试和调试

---

## 下一步行动

**请选择**：

### 选项 1：实现 MVP（推荐）
我将创建以下文件：
1. `internal/data/data.go` - 数据层初始化
2. `internal/data/campaign.go` - Campaign Repository
3. `internal/biz/campaign.go` - Campaign 业务逻辑
4. `api/marketing/v1/campaign.proto` - Campaign API
5. `internal/service/campaign.go` - Campaign Service
6. 配置 Wire 依赖注入

### 选项 2：只实现数据层
专注于完成所有 Repository 的实现

### 选项 3：自定义
您指定想要实现的部分

**您希望我按照哪个选项继续？**
