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

### 7. Monitoring and Observability

#### Logging
- Cache hits/misses are logged for monitoring
- Cache invalidation operations are logged
- Error conditions are properly logged with context

#### Key Patterns for Monitoring
- Monitor cache hit rates for performance insights
- Track invalidation patterns for optimization opportunities
- Monitor error rates for reliability assessment

## Usage Examples

### Plan Service with Caching
```go
// Cache-aside read
plan, err := planService.GetSubscriptionPlan(ctx, planID)

// Write-through create with invalidation
plan, err := planService.CreateSubscriptionPlan(ctx, creatorID, req)

// Update with selective invalidation  
plan, err := planService.UpdateSubscriptionPlan(ctx, planID, updateReq)
```

### User Subscription Service with Caching
```go
// Cached active subscriptions
subscriptions, err := userSubService.GetUserActiveSubscriptions(ctx, userID)

// Write-through create
subscription, err := userSubService.CreateUserSubscription(ctx, req)

// Selective invalidation for frequent operations
err := userSubService.UpdateTrafficUsage(ctx, subscriptionID, usedBytes)
```

### Order Service with Caching
```go
// Cached order retrieval
order, err := orderService.GetSubscriptionOrder(ctx, orderID)

// Order creation with caching
response, err := orderService.CreateSubscriptionOrder(ctx, req)

// Payment processing with cache invalidation
err := orderService.ProcessOrderPaymentSuccess(ctx, orderID)
```

## Integration with Existing Infrastructure

### Module Registration
The cached services are registered in the subscription domain module (`module.go`) as replacements for the base services, ensuring seamless integration with the existing dependency injection system.

### Cache Infrastructure Usage
- Uses the existing Redis cache infrastructure from `/internal/shared/cache/`
- Leverages predefined cache keys from `AllCacheKeys.Subscription`
- Follows established TTL constants (ShortCacheTTL, MediumCacheTTL, LongCacheTTL)

### Backward Compatibility
- All existing interfaces are maintained
- No breaking changes to service contracts
- Transparent caching that doesn't affect calling code

## Performance Benefits

### Expected Improvements
- **Plan retrieval**: ~80% reduction in database load for frequently accessed plans
- **User subscriptions**: ~60% reduction in database load for active user queries
- **Order lookups**: ~50% reduction in database load for order status checks
- **List operations**: Significant improvement for paginated results

### Scalability Benefits
- Reduced database connection pressure
- Better response times for high-traffic operations
- Improved concurrent user handling capability

## Maintenance Guidelines

### Cache Key Management
- Use consistent naming patterns for cache keys
- Include version information when cache structure changes
- Monitor cache key space usage and cleanup old patterns

### TTL Tuning
- Monitor cache hit rates and adjust TTLs as needed
- Consider business requirements when setting cache durations
- Balance consistency needs vs. performance gains

### Invalidation Optimization
- Review invalidation patterns regularly for efficiency
- Minimize unnecessary broad invalidations
- Consider implementing cache warming for critical data

This caching implementation provides a robust, scalable, and maintainable solution for the subscription domain while ensuring data consistency and optimal performance.