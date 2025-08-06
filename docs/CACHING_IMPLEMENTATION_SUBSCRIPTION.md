# 订阅领域缓存实现

## 概述

本文档描述了 Linke 项目中订阅领域的全面缓存实现。缓存层通过适当的缓存失效策略在保持数据一致性的同时提供显著的性能改进。

## 实现详情

### 1. 缓存架构

缓存实现遵循以下关键模式：
- **Cache-aside 模式** 用于实体读取
- **Write-through 模式** 用于创建/更新  
- **适当的缓存失效** 在更新/删除时
- **基于 TTL 的过期** 具有服务特定持续时间

### 2. 缓存 TTL 策略

| 服务 | TTL | 理由 |
|---------|-----|-----------|
| SubscriptionPlanService | 1 小时 (LongCacheTTL) | 套餐很少变化，可以安全地缓存更长时间 |
| UserSubscriptionService | 15 分钟 (MediumCacheTTL) | 用户订阅很关键，中等 TTL |
| SubscriptionOrderService | 15 分钟 (MediumCacheTTL) | 订单需要一致性，中等 TTL |

### 3. 缓存服务

#### SubscriptionPlanService (`subscription_plan_cached.go`)

**缓存方法：**
- `CreateSubscriptionPlan` - Write-through caching + list cache invalidation
- `GetSubscriptionPlan` - Cache-aside by ID
- `GetSubscriptionPlanByCode` - Cache-aside by code
- `GetSubscriptionPlans` - Complex query result caching
- `GetVisibleSubscriptionPlans` - Long TTL for public data
- `GetPopularSubscriptionPlans` - Long TTL for popular plans
- `UpdateSubscriptionPlan` - Write-through + selective invalidation
- `DeleteSubscriptionPlan` - Complete cache invalidation
- `ToggleSubscriptionPlanStatus` - Write-through + invalidation
- `ArchiveSubscriptionPlan` - Complete cache invalidation

**使用的缓存键：**
- `plan:id:{planID}` - Individual plan by ID
- `plan:code:{code}` - Individual plan by code
- `plan:list:*` - Plan list queries
- `plan:active:*` - Active plans
- `plan:popular:*` - Popular plans

#### UserSubscriptionService (`user_subscription_cached.go`)

**缓存方法：**
- `CreateUserSubscription` - Write-through + user cache invalidation
- `GetUserSubscription` - Cache-aside by ID
- `GetUserSubscriptionWithRelations` - Shorter TTL for relations
- `GetUserSubscriptions` - List result caching
- `GetUserSubscriptionsWithRelations` - Relations with short TTL
- `GetActiveUserSubscription` - Cache active subscriptions
- `GetUserActiveSubscriptions` - User-specific active subscriptions
- `UpdateUserSubscription` - Write-through + selective invalidation
- `CancelUserSubscription` - Cache invalidation
- `RenewUserSubscription` - Cache invalidation
- `DeleteUserSubscription` - Complete cache invalidation
- `UpdateLastUsed` - Selective cache invalidation (frequent operation)
- `UpdateTrafficUsage` - Selective cache invalidation (frequent operation)
- All traffic and renewal methods with appropriate invalidation
- Statistics methods with medium TTL

**使用的缓存键：**
- `subscription:id:{subscriptionID}` - Individual subscription by ID
- `subscription:user:{userID}` - User subscriptions
- `subscription:user:{userID}:active` - User active subscriptions
- `subscription:active:user:{userID}:plan:{planID}` - Active subscription for user/plan
- `subscription:list:*` - Subscription list queries

#### SubscriptionOrderService (`subscription_order_cached.go`)

**缓存方法：**
- `CreateSubscriptionOrder` - Write-through + user cache invalidation
- `GetSubscriptionOrder` - Cache-aside by ID
- `GetSubscriptionOrderByNumber` - Cache-aside by order number
- `GetSubscriptionOrders` - List result caching
- `GetUserSubscriptionOrders` - User-specific order caching
- `ProcessOrderPaymentSuccess` - Cache invalidation
- `CancelSubscriptionOrder` - Cache invalidation
- `GetOrderStatistics` - Statistics caching

**使用的缓存键：**
- `subscription:order:id:{orderID}` - Individual order by ID
- `subscription:order:number:{orderNumber}` - Order by number
- `subscription:user:{userID}:orders` - User orders
- `subscription:list:order:*` - Order list queries
- `subscription:stats:order:*` - Order statistics

### 4. 缓存失效策略

#### 粒度化失效
- **单个实体**: 按 ID 失效特定实体缓存
- **相关实体**: 当用户数据变化时失效用户特定缓存
- **列表缓存**: 失效相关的列表/查询结果缓存

#### 基于模式的失效
- 使用 Redis 模式匹配进行批量失效
- 按缓存键模式失效（例如，`plan:list:*`）
- 用户特定模式失效（例如，`subscription:user:{userID}:*`）

#### 选择性 vs. 广泛失效
- **频繁操作**（UpdateLastUsed、UpdateTrafficUsage）：选择性失效
- **关键操作**（创建、更新、删除）：全面失效
- **批量操作**（ProcessAutoRenewals）：广泛模式失效

### 5. 性能优化

#### 智能缓存决策
- **稳定数据的长 TTL**: 套餐、热门套餐（1小时）
- **业务数据的中等 TTL**: 订阅、订单（15分钟）
- **关联数据的短 TTL**: 带联结的数据（1分钟）

#### 高效缓存使用
- **批量操作**: 多个实体的 GetMany 操作
- **复杂查询**: 为过滤列表缓存整个结果集
- **统计数据**: 使用适当 TTL 缓存计算统计

#### 最小缓存开销
- **JSON 序列化**: 高效的序列化/反序列化
- **错误处理**: 缓存失败时优雅地回退到数据库
- **选择性失效**: 只失效必要的内容

### 6. 错误处理和可靠性

#### 缓存失败弹性
- 所有缓存操作都有数据库回退
- 缓存错误被记录但不会破坏功能
- 如果 Redis 不可用，服务继续运行

#### 数据一致性
- Write-through 模式确保缓存-数据库一致性
- 适当的失效防止过期数据
- 通过适当的锁定防止竞态条件

### 7. 监控和可观测性

#### 日志记录
- 记录缓存命中/未命中情况以供监控
- 记录缓存失效操作
- 错误情况会被适当记录并包含上下文信息

#### 监控的关键模式
- 监控缓存命中率以获取性能洞察
- 跟踪失效模式以寻找优化机会
- 监控错误率以评估可靠性

## 使用示例

### 套餐服务缓存
```go
// Cache-aside 读取
plan, err := planService.GetSubscriptionPlan(ctx, planID)

// Write-through 创建并失效
plan, err := planService.CreateSubscriptionPlan(ctx, creatorID, req)

// 选择性失效的更新
plan, err := planService.UpdateSubscriptionPlan(ctx, planID, updateReq)
```

### 用户订阅服务缓存
```go
// 缓存的活跃订阅
subscriptions, err := userSubService.GetUserActiveSubscriptions(ctx, userID)

// Write-through 创建
subscription, err := userSubService.CreateUserSubscription(ctx, req)

// 频繁操作的选择性失效
err := userSubService.UpdateTrafficUsage(ctx, subscriptionID, usedBytes)
```

### 订单服务缓存
```go
// 缓存的订单检索
order, err := orderService.GetSubscriptionOrder(ctx, orderID)

// 带缓存的订单创建
response, err := orderService.CreateSubscriptionOrder(ctx, req)

// 支付处理时的缓存失效
err := orderService.ProcessOrderPaymentSuccess(ctx, orderID)
```

## 与现有基础设施的集成

### 模块注册
缓存服务在订阅领域模块（`module.go`）中注册，作为基础服务的替代品，确保与现有的依赖注入系统无缝集成。

### 缓存基础设施使用
- 使用来自 `/internal/shared/cache/` 的现有 Redis 缓存基础设施
- 利用来自 `AllCacheKeys.Subscription` 的预定义缓存键
- 遵循已建立的 TTL 常量（ShortCacheTTL、MediumCacheTTL、LongCacheTTL）

### 向后兼容性
- 保持所有现有接口
- 不对服务契约进行破坏性更改
- 透明缓存不影响调用代码

## 性能优势

### 预期改进
- **套餐检索**: 频繁访问的套餐数据库负载减少约80%
- **用户订阅**: 活跃用户查询数据库负载减少约60%
- **订单查找**: 订单状态检查数据库负载减少约50%
- **列表操作**: 分页结果的显著改进

### 可扩展性优势
- 减少数据库连接压力
- 高流量操作的更好响应时间
- 提高并发用户处理能力

## 维护指南

### 缓存键管理
- 对缓存键使用一致的命名模式
- 当缓存结构变化时包含版本信息
- 监控缓存键空间使用情况并清理旧模式

### TTL 调优
- 监控缓存命中率并根据需要调整 TTL
- 设置缓存持续时间时考虑业务需求
- 平衡一致性需求与性能收益

### 失效优化
- 定期检查失效模式的效率
- 最小化不必要的广泛失效
- 考虑为关键数据实现缓存预热

此缓存实现为订阅领域提供了一个强大、可扩展且可维护的解决方案，同时确保数据一致性和最佳性能。