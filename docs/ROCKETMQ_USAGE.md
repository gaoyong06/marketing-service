# RocketMQ 使用指南

## 📦 依赖

已安装 Apache RocketMQ 官方 Go 客户端：

```bash
go get github.com/apache/rocketmq-client-go/v2
```

## 🔧 配置方式

### 方式一：不使用 RocketMQ（默认）

在 `wire_gen.go` 中，`NewTaskTriggerService` 的最后一个参数传入 `nil`：

```go
taskTriggerService := biz.NewTaskTriggerService(
    taskUseCase,
    taskCompletionLogUseCase,
    rewardGrantUseCase,
    rewardUseCase,
    campaignUseCase,
    inventoryReservationUseCase,
    validatorService,
    generatorService,
    distributorService,
    nil, // RocketMQ Producer 为 nil，不会发送消息
    logger,
)
```

### 方式二：使用 RocketMQ

在 `wire.go` 中创建 RocketMQ Producer，然后传入 `NewTaskTriggerService`：

```go
// 创建 RocketMQ Producer
rmqProducer, err := rocketmq.NewProducer(
    producer.WithNameServer([]string{"127.0.0.1:9876"}), // NameServer 地址
    producer.WithGroupName("marketing-producer-group"),  // Producer Group
    producer.WithRetry(2),
)
if err != nil {
    return nil, nil, err
}

if err := rmqProducer.Start(); err != nil {
    return nil, nil, err
}

// 使用 Producer 创建 TaskTriggerService
taskTriggerService := biz.NewTaskTriggerService(
    // ... 其他参数 ...
    rmqProducer, // 传入 RocketMQ Producer
    logger,
)
```

## 📝 使用示例

### 发布事件

`TaskTriggerService` 会自动在任务完成时发布事件到 RocketMQ：

```go
// 在 task_trigger.go 中，任务完成时会自动发布事件
eventMessage := &TaskEventMessage{
    EventType:    "USER_REGISTER",
    UserID:       123,
    TenantID:     "tenant1",
    AppID:        "app1",
    CampaignID:   "campaign-1",
    CampaignName: "测试活动",
    EventData:    map[string]interface{}{"count": 1},
    Timestamp:    time.Now().Format("2006-01-02T15:04:05Z07:00"),
}

// 发送到 RocketMQ Topic: marketing.task.completed
msg := primitive.NewMessage("marketing.task.completed", eventJSON)
result, err := rmqProducer.SendSync(ctx, msg)
```

### 订阅事件（在其他服务中）

```go
import (
    "github.com/apache/rocketmq-client-go/v2"
    "github.com/apache/rocketmq-client-go/v2/consumer"
)

// 创建 Consumer
consumer, err := rocketmq.NewPushConsumer(
    consumer.WithNameServer([]string{"127.0.0.1:9876"}),
    consumer.WithGroupName("marketing-consumer-group"),
    consumer.WithConsumerModel(consumer.Clustering),
)
if err != nil {
    panic(err)
}

// 订阅 Topic
err = consumer.Subscribe("marketing.task.completed", consumer.MessageSelector{}, 
    func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
        for _, msg := range msgs {
            // 处理消息
            var eventMessage biz.TaskEventMessage
            if err := json.Unmarshal(msg.Body, &eventMessage); err != nil {
                return consumer.ConsumeRetryLater, err
            }
            
            // 业务处理
            // ...
        }
        return consumer.ConsumeSuccess, nil
    },
)
```

## 🔗 相关链接

- [Apache RocketMQ 官方文档](https://rocketmq.apache.org/)
- [RocketMQ Go 客户端 GitHub](https://github.com/apache/rocketmq-client-go)
- [RocketMQ Go 客户端文档](https://github.com/apache/rocketmq-client-go/blob/master/README.md)

