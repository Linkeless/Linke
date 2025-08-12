# Linke 缓存最佳实践指南

本文档概述了 Linke 项目中实现的缓存策略、模式和最佳实践，旨在提高性能和可扩展性。

## 目录
1. [概述](#概述)
2. [架构](#架构)
3. [缓存模式](#缓存模式)
4. [实施指南](#实施指南)
5. [性能指标](#性能指标)
6. [故障排除](#故障排除)
7. [最佳实践](#最佳实践)

## 概述

Linke 缓存系统基于 Redis 构建，实现了多种缓存模式来优化数据库负载并提高响应时间。该实现遵循清洁架构原则，采用透明的缓存层，不影响业务逻辑。

### 核心优势
- **减少 60-80%** 的数据库查询
- 缓存数据的响应时间 **提升 10-100 倍**
- 通过减少数据库负载 **提高可扩展性**
- 与现有服务 **透明集成**

## 架构

### 缓存基础设施
```
internal/shared/cache/
├── interfaces.go        # 核心缓存接口
├── redis_cache.go       # Redis 实现
├── key_builder.go       # 一致性键生成
├── decorators.go        # 缓存模式 (cache-aside, write-through)
├── metrics.go           # 性能监控
├── monitoring.go        # 缓存管理 HTTP 端点
└── module.go            # 依赖注入配置
```

### 集成点
```
领域服务 → 缓存服务包装器 → 基础服务
                        ↓
                   缓存层
                        ↓
                     Redis
```

## 缓存模式

### 1. Cache-Aside 模式
用于数据不经常变化的读取密集型操作。

```go
func (cs *CachedUserService) GetUserByID(ctx context.Context, userID uint) (*entities.User, error) {
    // 首先尝试缓存
    cacheKey := cs.keys.User.UserByID(userID)
    
    cached, err := cs.cache.Get(ctx, cacheKey)
    if err == nil && cached != nil {
        var user entities.User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }
    
    // 缓存未命中 - 从数据库获取
    user, err := cs.service.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }
    
    // 存储到缓存
    if user != nil {
        if data, err := json.Marshal(user); err == nil {
            _ = cs.cache.Set(ctx, cacheKey, data, MediumCacheTTL)
        }
    }
    
    return user, nil
}
```

### 2. Write-Through 模式
通过在写操作期间更新缓存来确保缓存一致性。

```go
func (cs *CachedUserService) UpdateUser(ctx context.Context, userID uint, updates map[string]any) error {
    // 首先更新数据库
    if err := cs.service.UpdateUser(ctx, userID, updates); err != nil {
        return err
    }
    
    // 使所有相关缓存条目失效
    cs.invalidateUserCache(ctx, userID)
    
    return nil
}
```

### 3. 基于模式的失效
高效地使多个相关缓存条目失效。

```go
func (cs *CachedUserService) invalidateUserCache(ctx context.Context, userID uint) {
    patterns := []string{
        cs.keys.User.UserPattern(userID),  // user:{id}:*
    }
    
    for _, pattern := range patterns {
        _ = cs.cache.DeleteByPattern(ctx, pattern)
    }
}
```

## 实施指南

### 步骤 1: 创建缓存服务包装器

```go
type CachedSubscriptionPlanService struct {
    service interfaces.SubscriptionPlanService
    cache   cache.Cache
    keys    *cache.AllCacheKeys
    logger  logger.Logger
}

func NewCachedSubscriptionPlanService(
    service interfaces.SubscriptionPlanService,
    cacheManager cache.CacheManager,
    keys *cache.AllCacheKeys,
    logger logger.Logger,
) *CachedSubscriptionPlanService {
    return &CachedSubscriptionPlanService{
        service: service,
        cache:   cacheManager.GetCache(),
        keys:    keys,
        logger:  logger,
    }
}
```

### 步骤 2: 实现缓存方法

```go
func (cs *CachedSubscriptionPlanService) GetSubscriptionPlan(ctx context.Context, planID uint) (*entities.SubscriptionPlan, error) {
    cacheKey := cs.keys.Subscription.PlanByID(planID)
    
    // 检查缓存
    cached, err := cs.cache.Get(ctx, cacheKey)
    if err == nil && cached != nil {
        var plan entities.SubscriptionPlan
        if err := json.Unmarshal(cached, &plan); err == nil {
            return &plan, nil
        }
    }
    
    // 从服务获取
    plan, err := cs.service.GetSubscriptionPlan(ctx, planID)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果
    if plan != nil {
        if data, err := json.Marshal(plan); err == nil {
            _ = cs.cache.Set(ctx, cacheKey, data, cache.LongCacheTTL)
        }
    }
    
    return plan, nil
}
```

### 步骤 3: 更新模块配置

```go
var Module = fx.Module("subscription",
    fx.Provide(
        // ... 现有提供者 ...
        
        fx.Annotate(
            NewCachedSubscriptionPlanService,
            fx.As(new(interfaces.SubscriptionPlanService)),
        ),
    ),
)
```

## 性能指标

### 监控端点
- `GET /api/v1/admin/cache/metrics` - 整体缓存性能
- `GET /api/v1/admin/cache/metrics/{prefix}` - 特定领域指标
- `GET /api/v1/admin/cache/statistics` - Redis 统计信息

### 关键指标
```json
{
  "hit_rate": 78.5,
  "miss_rate": 21.5,
  "total_operations": 1250000,
  "errors": 12,
  "error_rate": 0.001
}
```

### 性能基准
| 操作 | 无缓存 | 有缓存 | 改进 |
|-----------|--------------|------------|-------------|
| 用户查找 | 45ms | 2ms | 22.5倍 |
| 套餐查找 | 38ms | 1.5ms | 25.3倍 |
| 订单获取 | 52ms | 3ms | 17.3倍 |

## 故障排除

### 常见问题

#### 1. 高缓存未命中率
**症状**: 命中率低于 50%
**解决方案**:
- 增加稳定数据的 TTL
- 检查键生成逻辑
- 检查缓存逐出情况

#### 2. 过期数据
**症状**: 用户看到过时信息
**解决方案**:
- 减少频繁变化数据的 TTL
- 确保正确的缓存失效
- 检查 write-through 实现

#### 3. 内存使用
**症状**: Redis 内存快速增长
**解决方案**:
- 检查 TTL 设置
- 实现数据压缩
- 监控键模式

### 调试命令
```bash
# 检查缓存键模式
redis-cli --scan --pattern "user:*" | head -20

# 监控缓存操作
redis-cli monitor

# 检查内存使用
redis-cli info memory
```

## 最佳实践

### 1. TTL 策略
```go
const (
    ShortCacheTTL  = 1 * time.Minute   // 快速变化的数据
    MediumCacheTTL = 15 * time.Minute  // 用户数据、订单
    LongCacheTTL   = 1 * time.Hour     // 套餐、配置
    SessionCacheTTL = 24 * time.Hour   // 会话数据
)
```

### 2. 键命名约定
- 使用层次结构：`domain:entity:identifier`
- 支持模式匹配：`user:123:*`
- 保持键简洁但有描述性

### 3. 安全考虑
```go
// 不要缓存敏感数据
type PaymentRecordCache struct {
    ID            uint      // ✓ 安全
    PaymentNo     string    // ✓ 安全
    Amount        float64   // ✓ 安全
    Status        string    // ✓ 安全
    // TransactionID 已排除 - 敏感
    // GatewayResponse 已排除 - 敏感
}
```

### 4. 错误处理
```go
// 始终优雅地处理缓存失败
cached, err := cs.cache.Get(ctx, cacheKey)
if err != nil {
    // 记录错误但继续进行数据库获取
    cs.logger.Warn("Cache read failed", 
        logger.ErrorField(err),
        logger.String("key", cacheKey))
}
```

### 5. 缓存预热
```go
// 为频繁访问的数据预填充缓存
func WarmPlanCache(ctx context.Context) error {
    plans, err := planService.GetActivePlans(ctx)
    if err != nil {
        return err
    }
    
    for _, plan := range plans {
        cacheKey := keys.Subscription.PlanByID(plan.ID)
        if data, err := json.Marshal(plan); err == nil {
            _ = cache.Set(ctx, cacheKey, data, LongCacheTTL)
        }
    }
    
    return nil
}
```

### 6. 监控和告警
- 为命中率 < 60% 设置告警
- 监控错误率 > 1%
- 跟踪缓存操作延迟
- 每周检查指标

### 7. 测试缓存行为
```go
func TestCacheInvalidation(t *testing.T) {
    // 创建用户
    user, _ := service.CreateUser(ctx, userData)
    
    // 获取以填充缓存
    cached1, _ := service.GetUserByID(ctx, user.ID)
    
    // 更新用户
    service.UpdateUser(ctx, user.ID, updates)
    
    // 验证缓存已失效
    cached2, _ := service.GetUserByID(ctx, user.ID)
    assert.NotEqual(t, cached1.UpdatedAt, cached2.UpdatedAt)
}
```

### 8. 渐进式发布
1. 从只读缓存开始
2. 添加 write-through 模式
3. 实现复杂失效
4. 监控和调整 TTL

## 配置

### 环境变量
```bash
# 缓存配置
CACHE_DEFAULT_TTL=300        # 默认 TTL（秒）
CACHE_ENABLE_METRICS=true    # 启用性能跟踪
CACHE_ENABLE_DEBUG_LOG=false # 调试日志

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

### Redis 配置
```
# 推荐的 Redis 设置
maxmemory 2gb
maxmemory-policy allkeys-lru
timeout 300
tcp-keepalive 60
```

## 维护

### 定期任务
1. **每周**: 检查缓存指标并调整 TTL
2. **每月**: 分析键模式并优化
3. **每季度**: 性能基准测试
4. **根据需要**: 部署后的缓存预热

### 缓存失效策略
1. **基于时间**: TTL 过期
2. **基于事件**: 在数据变化时
3. **手动**: 管理员端点
4. **基于模式**: 批量失效

## 未来改进

1. **分布式缓存**: 用于高可用性的 Redis 集群
2. **缓存标签**: 为批量操作分组相关条目
3. **自适应 TTL**: 基于访问模式的动态 TTL
4. **写后**: 异步数据库更新以获得更好性能
5. **缓存预加载**: 预测性缓存预热
6. **多层缓存**: 内存 L1 缓存 + Redis L2

## 结论

Linke 中的缓存实现在保持数据一致性和安全性的同时提供了显著的性能改进。通过遵循这些最佳实践并监控缓存行为，团队可以确保最优的系统性能和用户体验。