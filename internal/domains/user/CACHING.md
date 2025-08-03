# 用户领域缓存实现

本文档描述了 Linke 项目中用户领域的全面缓存实现。

## 概述

用户领域现在包含一个复杂的缓存层，实现了多种缓存模式以提高性能并减少数据库负载。该实现使用 `CachedUserService` 包装现有的 `UserService`，在不改变服务接口的情况下提供透明的缓存。

## 架构

### 组件

- **CachedUserService**: 主要服务，使用缓存逻辑包装基础 UserService
- **Cache Manager**: 管理基于 Redis 的缓存操作
- **Cache Keys**: 结构化缓存键生成，确保命名一致性
- **TTL Configuration**: 不同缓存条目的可配置生存时间

### 缓存键结构

用户数据使用以下缓存键：

- `user:id:{user_id}` - 按 ID 缓存用户
- `user:email:{email}` - 按邮箱地址缓存用户
- `user:username:{username}` - 按用户名缓存用户（如果可用）
- `user:profile:{user_id}` - 缓存用户资料数据
- `user:{user_id}:*` - 用于失效所有用户相关缓存条目的模式

## 已实现的缓存模式

### 1. Cache-Aside 模式

用于读取操作（`GetUserByID`、`GetUserByEmail`、`GetActiveUserByID`、`GetActiveUserByEmail`）：

```go
// 1. 首先检查缓存
if user, err := s.getUserFromCache(ctx, cacheKey); err == nil && user != nil {
    return user, nil
}

// 2. 缓存未命中 - 从数据库获取
user, err := s.baseService.GetUserByID(ctx, id)
if err != nil {
    return nil, err
}

// 3. 存储到缓存以备将来请求
s.setUserInCache(ctx, cacheKey, user, cache.MediumCacheTTL)
```

### 2. Write-Through 模式

用于创建和更新操作（`CreateUser`、`UpdateUser`）：

```go
// 1. 首先执行数据库操作
if err := s.baseService.CreateUser(ctx, user); err != nil {
    return err
}

// 2. 立即存储到缓存
s.setUserInCache(ctx, keyByID, user, ttl)
s.setUserInCache(ctx, keyByEmail, user, ttl)
```

### 3. 缓存失效

用于修改或删除用户的操作：

- **立即失效**: 对于删除操作，缓存立即失效
- **基于模式的失效**: 使用通配符模式失效所有相关缓存条目
- **多键失效**: 同时失效 ID、邮箱和用户名的缓存

## TTL（生存时间）配置

| 缓存类型 | TTL | 理由 |
|----------|-----|------|
| 用户数据 | 15分钟（`MediumCacheTTL`） | 平衡性能与数据新鲜度 |
| 用户资料 | 15分钟 | 资料数据变化频率较低 |
| 活跃用户查找 | 15分钟 | 状态变化需要合理的传播时间 |

## 方法和缓存行为

### 缓存方法

| 方法 | 缓存模式 | 使用的键 | TTL |
|------|----------|----------|-----|
| `CreateUser` | Write-Through | ID, Email, Username | 15分钟 |
| `GetUserByID` | Cache-Aside | ID | 15分钟 |
| `GetUserByEmail` | Cache-Aside | Email | 15分钟 |
| `GetActiveUserByID` | Cache-Aside（带验证） | ID | 15分钟 |
| `GetActiveUserByEmail` | Cache-Aside（带验证） | Email | 15分钟 |
| `UpdateUser` | Write-Through + 失效 | ID, Email, Username | 15分钟 |
| `UpdateUserStatus` | Write-Through + 失效 | ID, Email, Username | 15分钟 |
| `UpdateUserRole` | Write-Through + 失效 | ID, Email, Username | 15分钟 |
| `SoftDeleteUser` | 仅失效 | 所有用户键 | - |
| `RestoreUser` | 仅失效 | 所有用户键 | - |
| `HardDeleteUser` | 仅失效 | 所有用户键 | - |
| `BatchDeleteUsers` | 仅失效 | 所有受影响的用户键 | - |
| `BatchRestoreUsers` | 仅失效 | 所有受影响的用户键 | - |

### 非缓存方法

由于其性质，以下方法**不被缓存**：

- `ListUsers` - 带分页的动态结果集
- `ListDeletedUsers` - 管理员操作，通常不是性能关键
- `ListUsersByProvider` - 具有可变结果的过滤操作
- `SearchUsers` - 带动态查询的搜索操作
- `GetUserStats` - 经常变化的统计数据

## 缓存失效策略

### 单用户失效

当用户被修改或删除时，以下缓存条目将失效：

1. `user:id:{user_id}`
2. `user:email:{email}`（如果邮箱存在）
3. `user:username:{username}`（如果用户名存在）
4. `user:profile:{user_id}`
5. 所有匹配 `user:{user_id}:*` 模式的条目

### 批量操作

对于批量操作，使用与单用户失效相同的策略为每个受影响的用户执行失效。

### 优雅降级

如果缓存操作失败：
- 读取操作回退到数据库查询
- 写入操作继续正常工作
- 缓存错误被记录但不影响业务逻辑

## 性能优势

### 预期的性能改进

1. **减少数据库负载**: 经常访问的用户从缓存中提供
2. **降低延迟**: Redis 访问通常比数据库查询快 10-100 倍
3. **更好的可扩展性**: 减少数据库连接和查询负载
4. **改善用户体验**: 更快的身份验证和用户查找操作

### 缓存命中率优化

此实现旨在实现以下场景的高缓存命中率：
- 用户身份验证流程（登录、JWT 验证）
- 用户资料访问
- 权限检查
- 其他服务中的频繁用户查找

## 配置

### 缓存配置

缓存行为通过 `CacheConfig` 结构进行配置：

```go
type CacheConfig struct {
    DefaultTTL      time.Duration  // 默认缓存 TTL
    MaxRetries      int           // 缓存操作的最大重试次数
    RetryDelay      time.Duration // 重试之间的延迟
    EnableMetrics   bool         // 启用缓存指标收集
    CompressionType string       // 大型缓存条目的压缩
}
```

### 环境变量

缓存行为可以通过以下变量影响：
- `REDIS_HOST`、`REDIS_PORT`、`REDIS_PASSWORD` - Redis 连接
- `CACHE_DEFAULT_TTL` - 缓存条目的默认 TTL
- `CACHE_ENABLE_METRICS` - 启用缓存指标收集

## 监控和可观测性

### 日志记录

所有缓存操作都包含结构化日志：
- 缓存命中和未命中
- 缓存失效事件
- 缓存操作错误
- 性能指标

### 指标

启用时，将收集以下指标：
- 缓存命中/未命中比率
- 缓存操作延迟
- 缓存条目数量
- 错误率

### 日志示例

```
INFO  User cache hit for ID lookup    user_id=123 cache_key=user:id:123
INFO  User created and cached         user_id=456 email=user@example.com
INFO  User cache invalidated          user_id=123
WARN  Cache operation failed          operation=get key=user:id:123 error="connection refused"
```

## 最佳实践

### 开发者

1. **始终使用注入的服务**: DI 容器默认提供缓存服务
2. **监控缓存命中率**: 使用日志和指标优化缓存使用
3. **意识到缓存一致性**: 理解缓存失效发生的时机
4. **在启用缓存的情况下测试**: 在集成测试中包含缓存测试

### 运维

1. **监控 Redis 健康状态**: 缓存故障不应影响核心功能
2. **设置适当的内存限制**: 配置 Redis 内存策略
3. **监控缓存命中率**: 低命中率可能表示使用不当
4. **定期缓存清理**: 确保过期条目被正确清理

## 测试

### 单元测试

实现包含以下的全面单元测试：
- 缓存命中/未命中场景
- 缓存失效逻辑
- 键生成一致性
- 错误处理和回退行为

### 集成测试

使用以下方式测试缓存层：
- 真实的 Redis 实例
- 缓存性能基准测试
- 错误场景（Redis 不可用）
- 跨操作的缓存一致性

### 运行测试

```bash
# 运行用户领域测试
go test ./internal/domains/user/...

# 运行缓存特定测试
go test ./internal/domains/user/usecases/implementations -run TestCached

# 运行缓存性能测试
go test -bench=. ./internal/domains/user/...
```

## 故障排除

### 常见问题

1. **新数据的缓存未命中**: 确保 write-through 模式正确实现
2. **缓存中的过期数据**: 检查更新操作的缓存失效逻辑
3. **内存使用**: 监控 Redis 内存使用情况并设置适当的 TTL
4. **连接问题**: 确保 Redis 连接并优雅地处理故障

### 调试命令

```bash
# 检查 Redis 连接
redis-cli ping

# 监控缓存操作
redis-cli monitor

# 检查缓存键模式
redis-cli keys "user:*"

# 检查特定用户缓存
redis-cli get "user:id:123"
```

## 未来增强

### 计划中的改进

1. **缓存预热**: 预加载经常访问的用户
2. **自适应 TTL**: 根据用户活动模式调整 TTL
3. **多级缓存**: 添加内存 L1 缓存以获得超快访问
4. **缓存标签**: 实现更复杂的失效策略
5. **指标仪表板**: 实时缓存性能监控

### 性能调优

1. **压缩**: 为大型用户对象启用压缩
2. **管道化**: 批量缓存操作以获得更好性能
3. **连接池**: 优化 Redis 连接管理
4. **异步失效**: 后台缓存失效以获得更好的响应时间

## 迁移和回退

### 启用缓存

缓存实现是透明的，可以通过依赖注入配置启用/禁用，而无需更改应用程序代码。

### 回退计划

如果缓存导致问题：
1. 更新 DI 配置以使用基础 `UserService` 而不是 `CachedUserService`
2. 无需数据迁移
3. 可以安全地清理缓存：`redis-cli flushdb`
4. 应用程序在没有缓存的情况下继续正常工作

此实现提供了一个强大、可扩展的缓存解决方案，在保持数据一致性和提供运维灵活性的同时显著改善了用户领域的性能。