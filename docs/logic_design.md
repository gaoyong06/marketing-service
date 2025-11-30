# 营销服务业务逻辑设计

本文档主要通过 UML 图表展示营销服务的核心业务逻辑和流程。

## 1. 核心领域模型 (Domain Model)

展示四个核心实体及其关系。Generator/Validator/Distributor 作为配置组件存储在 Reward 表中。

```mermaid
classDiagram
    class Campaign {
        +String CampaignID
        +String Type
        +Time StartTime
        +Time EndTime
        +Status Status
        +JSON AudienceConfig
        +JSON ValidatorConfig
    }

    class Audience {
        +String AudienceID
        +String Type
        +JSON RuleConfig
    }

    class Task {
        +String TaskID
        +String Type
        +JSON TriggerConfig
        +JSON ConditionConfig
        +String RewardID
    }

    class Reward {
        +String RewardID
        +String Type
        +JSON ContentConfig
        +JSON GeneratorConfig
        +JSON ValidatorConfig
        +JSON DistributorConfig
        +Int Version
    }

    class RewardGrant {
        +String GrantID
        +String RewardID
        +String RewardName (冗余)
        +String CampaignName (冗余)
        +String TaskName (冗余)
        +JSON ContentSnapshot
        +Status Status
    }

    Campaign "1" *-- "N" Task : 组合
    Campaign "1" o-- "1" Audience : 引用
    Task "1" o-- "1" Reward : 关联
    Reward "1" -- "N" RewardGrant : 发放
```

**设计说明**：
- **核心实体**：Campaign, Audience, Task, Reward 存储在独立表中
- **配置组件**：Generator/Validator/Distributor 以 JSON 格式存储在 Reward 表中
- **冗余字段**：RewardGrant 中的 name 字段采用快照模式，不随源数据修改


## 2. 活动创建流程 (Campaign Creation)

展示如何通过积木组合创建一个活动。

```mermaid
sequenceDiagram
    participant Admin as 运营人员
    participant API as Marketing API
    participant DB as Database

    Admin->>API: 1. 创建活动 (CreateCampaign)
    API->>DB: Insert Campaign
    DB-->>API: CampaignID
    API-->>Admin: Success

    Admin->>API: 2. 关联受众 (AddAudience)
    API->>DB: Insert Campaign_Audience
    API-->>Admin: Success

    Admin->>API: 3. 添加任务 (AddTask)
    Note right of Admin: 定义触发器(Trigger)和条件(Condition)
    API->>DB: Insert Task
    API->>DB: Insert Campaign_Task
    API-->>Admin: Success

    Admin->>API: 4. 配置奖励 (AddReward)
    Note right of Admin: 配置JSON：generator_config, validator_config, distributor_config
    API->>DB: Insert Reward (包含JSON配置)
    API-->>Admin: Success

    Admin->>API: 5. 发布活动 (PublishCampaign)
    API->>DB: Update Campaign Status = ACTIVE
    API-->>Admin: Success
```

## 3. 任务触发与完成流程 (Task Trigger & Completion)

展示基于事件驱动的任务处理流程。

```mermaid
sequenceDiagram
    participant User as 用户
    participant EventBus as 事件总线 (Kafka/RocketMQ)
    participant TaskSvc as 任务服务
    participant RuleEngine as 规则引擎
    participant DB as Database
    participant RewardSvc as 奖励服务

    User->>EventBus: 产生业务事件 (如: 支付成功)
    EventBus->>TaskSvc: 消费事件 (Trigger匹配)
    
    TaskSvc->>DB: 查询活跃活动 & 任务 (ListActiveTasks)
    DB-->>TaskSvc: 任务列表 (含Trigger配置)

    loop 遍历任务
        TaskSvc->>TaskSvc: 匹配 Trigger (Event Type & Condition)
        
        alt Trigger 匹配成功
            TaskSvc->>RuleEngine: 校验完成条件 (Check Condition)
            RuleEngine-->>TaskSvc: Pass/Fail
            
            alt 条件满足
                TaskSvc->>DB: 记录任务完成 (Insert TaskCompletionLog)
                TaskSvc->>RewardSvc: 触发奖励发放 (IssueReward)
            end
        end
    end
```

## 4. 奖励发放流程 (Reward Distribution)

展示奖励发放的详细逻辑，包括校验、库存扣减和实际发放。

```mermaid
sequenceDiagram
    participant Client as 调用方 (Task/API)
    participant RewardSvc as 奖励服务
    participant Validator as 校验器 (Validator)
    participant Inventory as 库存服务 (Inventory)
    participant Generator as 生成器 (Generator)
    participant Distributor as 发放器 (Distributor)
    participant DB as Database

    Client->>RewardSvc: 请求发放奖励 (IssueReward)
    
    %% 1. 校验阶段
    RewardSvc->>Validator: 执行校验链 (Check Validators)
    Note right of Validator: 1. 时间校验<br/>2. 资格校验(Audience)<br/>3. 频次限制(Limit)
    Validator-->>RewardSvc: Pass/Fail
    
    alt 校验不通过
        RewardSvc-->>Client: Error (校验失败)
    end

    %% 2. 库存预占
    RewardSvc->>Inventory: 预占库存 (Reserve)
    Inventory-->>RewardSvc: Success/Fail
    
    alt 库存不足
        RewardSvc-->>Client: Error (库存不足)
    end

    %% 3. 生成奖励内容
    RewardSvc->>Generator: 生成奖励内容 (Generate)
    Note right of Generator: 生成兑换码 / 计算积分 / 获取券码
    Generator-->>RewardSvc: Content

    %% 4. 持久化发放记录
    RewardSvc->>DB: 创建奖励发放记录 (Insert RewardGrant)
    Note right of DB: 快照模式：记录campaign_name, reward_name等
    DB-->>RewardSvc: GrantID

    %% 5. 实际发放
    RewardSvc->>Distributor: 执行发放 (Distribute)
    Note right of Distributor: 调用下游服务 / 发邮件 / 发短信
    Distributor-->>RewardSvc: Success

    %% 6. 更新状态
    RewardSvc->>DB: 更新实例状态 (DISTRIBUTED)
    RewardSvc->>Inventory: 确认扣减 (Commit)
    
    RewardSvc-->>Client: Success (发放成功)
```
