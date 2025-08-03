# 事件驱动架构实现

本文档描述了为 Linke 项目实现的综合事件驱动架构。

## 概述

事件驱动架构为跨领域处理业务事件提供了一个强大、可扩展且可维护的系统。它支持同步和异步事件处理，包含事件溯源功能，并实现了不同业务领域之间的松耦合。

## 架构组件

### 1. 核心事件基础设施（`/internal/shared/events/`）

#### 事件接口和基础类型
- **事件接口**: 定义所有事件的契约
- **BaseEvent**: 提供通用事件功能
- **特定领域事件**: UserEvent、PaymentEvent、SubscriptionEvent、OrderEvent、InvoiceEvent、ServerEvent

#### 事件总线和发布器
- **InMemoryEventBus**: 应用程序内的同步事件处理
- **EnhancedEventBus**: 具有订阅者管理和统计功能的高级事件总线
- **AsyncEventBus**: 使用队列系统的异步事件处理

#### 事件存储
- **DatabaseEventStore**: 为审计跟踪和事件源持久化事件
- **事件重放**: 从特定时间戳重放事件的能力
- **事件统计**: 关于存储事件的综合指标

### 2. 领域事件

#### 用户事件
- `user.created` - 创建新用户时
- `user.registered` - 用户完成注册时
- `user.updated` - 修改用户信息时
- `user.deleted` - 软/硬删除用户时
- `user.status_changed` - 用户状态变化时（活跃、非活跃、被禁止）
- `user.logged_in` - 用户登录时
- `user.logged_out` - 用户登出时
- `user.password_reset` - 重置密码时

#### 支付事件
- `payment.created` - 发起支付时
- `payment.completed` - 支付成功时
- `payment.failed` - 支付失败时
- `payment.refunded` - 支付退款时

#### 订阅事件
- `subscription.created` - 创建订阅时
- `subscription.activated` - 订阅变为活跃时
- `subscription.expired` - 订阅过期时
- `subscription.cancelled` - 取消订阅时
- `subscription.renewed` - 续订时
- `subscription.suspended` - 暂停订阅时

#### 订单事件
- `order.created` - 创建订单时
- `order.updated` - 修改订单详情时
- `order.paid` - 支付订单时
- `order.cancelled` - 取消订单时
- `order.expired` - 订单过期时
- `order.refunded` - 订单退款时

#### 发票事件
- `invoice.created` - 创建发票时
- `invoice.generated` - 生成发票时
- `invoice.sent` - 发送发票时
- `invoice.paid` - 支付发票时
- `invoice.overdue` - 发票逾期时
- `invoice.cancelled` - 取消发票时

### 3. 跨领域事件处理器

#### 支付 → 订单 → 订阅 流程
```
支付完成 → 订单已付 → 发票已创建 + 订阅已激活
```

#### 订阅管理
```
订阅过期 → 用户状态变更
发票逾期 → 订阅暂停
```

#### 用户生命周期
```
用户删除 → 订阅取消（所有活跃订阅）
```

#### 支付失败
```
支付失败 → 订单取消
```

### 4. 异步处理

#### 队列集成
- 可以使用现有的 Redis/Asynq 队列系统异步处理事件
- 失败事件处理的重试机制
- 所有重试都失败的事件的死信队列

#### 任务类型
- `event:process` - 异步处理事件
- `event:reprocess` - 重新处理失败事件

### 5. Event Store Schema

```sql
CREATE TABLE event_store (
    id BIGSERIAL PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    event_source VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(100),
    aggregate_type VARCHAR(50),
    event_version VARCHAR(20) NOT NULL DEFAULT '1.0',
    event_data TEXT NOT NULL,
    metadata TEXT,
    occurred_at TIMESTAMP WITH TIME ZONE NOT NULL,
    stored_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
```

## 使用示例

### 1. 在领域服务中发布事件

```go
// 在 UserService 中
func (s *EventAwareUserService) CreateUser(ctx context.Context, user *entities.User) error {
    // 在数据库中创建用户
    if err := s.userService.CreateUser(ctx, user); err != nil {
        return err
    }

    // 发布用户创建事件
    userEvent := events.NewUserEvent(
        events.EventTypeUserCreated,
        user.ID,
        map[string]interface{}{
            "user_id": user.ID,
            "email":   user.Email,
            "name":    user.Name,
            // ... 其他相关数据
        },
    )

    // 同步发布
    if err := s.eventBus.Publish(ctx, userEvent); err != nil {
        s.logger.Error("Failed to publish user created event", logger.ErrorField(err))
        // 事件发布失败不会导致操作失败
    }

    return nil
}
```

### 2. 订阅事件

```go
// 注册事件处理器
func RegisterEventHandlers(eventBus events.EventBus) {
    // 创建通知处理器
    notificationHandler := NewNotificationHandler()
    
    // 订阅相关事件
    eventBus.Subscribe(notificationHandler.EventTypes(), notificationHandler)
}
```

### 3. 跨领域事件处理

```go
// 支付完成触发订单已付事件
func (h *CrossDomainEventHandlers) PaymentCompletedHandler() events.EventHandler {
    return events.NewEventHandler(
        []string{events.EventTypePaymentCompleted},
        func(ctx context.Context, event events.Event) error {
            paymentEvent := event.(*events.PaymentEvent)
            
            // 从支付数据中提取订单ID
            if orderData, ok := paymentEvent.EventData().(map[string]interface{}); ok {
                if orderID, exists := orderData["order_id"]; exists {
                    // 创建并发布订单已付事件
                    orderPaidEvent := events.NewOrderEvent(
                        events.EventTypeOrderPaid,
                        uint(orderID.(float64)),
                        paymentEvent.UserID,
                        map[string]interface{}{
                            "payment_id": paymentEvent.PaymentID,
                            "amount":     paymentEvent.Amount,
                        },
                    )
                    
                    return events.Publish(ctx, orderPaidEvent)
                }
            }
            
            return nil
        },
    )
}
```

### 4. 异步事件处理

```go
// 异步发布事件
asyncEventBus := events.NewAsyncEventBus(eventBus, asyncProcessor)
err := asyncEventBus.PublishAsync(ctx, event)
```

### 5. 事件重放

```go
// 从特定时间戳重放事件
replayHandler := events.NewEventReplayHandler(eventStore, asyncProcessor)
err := replayHandler.ReplayEvents(ctx, fromTimestamp, []string{events.EventTypeUserCreated})
```

## 配置

### 模块设置

事件系统在共享模块（`/internal/shared/module.go`）中自动配置：

```go
// Event system providers
fx.Provide(
    // Event store
    func(db *gorm.DB) events.EventStore {
        return events.NewDatabaseEventStore(db)
    },
    // Event bus
    func() events.EventBus {
        return events.NewEnhancedEventBus()
    },
    // Async event processor
    func(taskQueue *queue.TaskQueue, eventStore events.EventStore, eventBus events.EventBus) *events.AsyncEventProcessor {
        return events.NewAsyncEventProcessor(taskQueue, eventStore, eventBus, events.DefaultRetryConfig())
    },
    // Cross-domain handlers
    func() *events.CrossDomainEventHandlers {
        return events.NewCrossDomainEventHandlers()
    },
),
```

### 领域集成

领域模块可以集成事件感知服务：

```go
// In user module
fx.Provide(
    fx.Annotate(
        func(cachedUserService *implementations.CachedUserService, eventBus events.EventBus, logger framework.Logger) interfaces.UserService {
            return implementations.NewEventAwareUserService(cachedUserService, eventBus, logger)
        },
        fx.As(new(interfaces.UserService)),
    ),
),
```

## 最佳实践

### 1. 事件设计
- 事件应该是不可变的
- 事件名称使用过去时（例如，“UserCreated” 而不是 “CreateUser”）
- 在事件中包含所有必要数据，以避免与当前状态耦合
- 使用相关ID来跟踪相关事件

### 2. 错误处理
- 事件发布失败不应导致主要操作失败
- 为异步事件处理使用重试机制
- 为无法处理的事件实现死信队列
- 记录所有事件处理错误

### 3. 性能
- 为非关键事件使用异步处理
- 尽可能批量处理事件
- 监控事件处理延迟和吞吐量
- 在事件存储中实现适当的索引

### 4. 测试
- 单独测试事件处理器
- 单元测试使用内存事件总线
- 端到端测试跨领域事件流
- 在事件处理器中模拟外部依赖

## 监控和可观测性

### 事件存储统计
```go
stats, err := eventStore.GetStats(ctx)
// 返回：总事件数、按类型/来源的事件、最近活动
```

### 订阅者健康检查
```go
health := enhancedEventBus.HealthCheck()
// 返回：所有订阅者的健康状态
```

### 事件处理指标
```go
metrics, err := asyncProcessor.GetMetrics(ctx)
// 返回：已处理/失败事件、队列长度、处理时间
```

## 数据库迁移

要启用事件存储，请运行迁移：

```bash
make migrate-up
```

这将创建带有适当索引的 `event_store` 表，用于高效查询。

## 未来增强

1. **Redis 事件总线**: 跨多个实例的分布式事件处理
2. **事件版本控制**: 处理事件模式演进
3. **事件快照**: 优化事件重放性能
4. **GraphQL 订阅**: 向客户端实时事件流
5. **事件源**: 关键聚合的完整事件源实现

## 故障排除

### 常见问题

1. **事件未被处理**
   - 检查处理器是否正确注册
   - 验证发布者和订阅者之间的事件类型匹配
   - 检查异步事件的任务处理器是否正在运行

2. **事件存储错误**
   - 确保数据库迁移已运行
   - 检查数据库权限
   - 验证事件序列化/反序列化

3. **高事件量的内存问题**
   - 对高量事件使用异步处理
   - 实现事件批处理
   - 考虑事件归档策略

### 调试命令

```bash
# 检查事件存储状态
make migrate-status

# 运行事件系统测试
go test ./internal/shared/events/...

# 检查队列状态（用于异步事件）
# （实现取决于您的监控设置）
```

## 结论

这个事件驱动架构为构建可扩展、可维护的应用程序提供了坚实的基础。它实现了领域之间的松耦合，支持同步和异步处理，通过事件源提供审计功能，并包含全面的测试和监控功能。