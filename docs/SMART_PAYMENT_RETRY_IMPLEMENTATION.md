# 智能支付重试策略实现

本文档详细介绍了 Linke 支付系统智能支付重试策略功能的全面实现，基于 `SUBSCRIPTION_OPTIMIZATION.md` 中概述的需求。

## 概述

智能支付重试策略使用可配置的重试策略、指数回退算法和失败类型分类，智能地重试失败的支付。该系统显著提高支付成功率，同时减少手动干预。

## 关键功能

### 1. 智能重试逻辑
- **指数回退**: 默认策略，延迟时间为 1 小时、6 小时和 24 小时
- **线性回退**: 特定网关的替代策略
- **自定义策略**: 网关特定的重试配置
- **失败分类**: 自动分类临时失败与永久失败

### 2. 可配置重试策略
- **按网关配置**: 不同支付网关使用不同策略
- **按支付方式**: 特定支付方式的精细化策略
- **动态配置**: 运行时更新，无需重启系统
- **默认回退**: 所有支持网关的全面默认策略

### 3. 全面监控
- **实时健康指标**: 系统级重试健康监控
- **网关特定统计**: 每个网关的成功率和性能指标
- **失败模式分析**: 自动检测常见失败模式
- **告警系统**: 系统健康问题的可配置告警

### 4. 管理员管理界面
- **重试队列管理**: 查看和管理所有活跃重试
- **批量操作**: 一次取消或重置多个重试
- **详细分析**: 全面的统计和报告
- **手动干预**: 管理员覆盖功能

## 架构

### 数据库架构

#### 支付重试表
```sql
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    attempt_number INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 3,
    next_retry_at TIMESTAMP NOT NULL,
    last_attempt_at TIMESTAMP NOT NULL,
    retry_strategy VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    failure_type VARCHAR(30) NULL,
    -- ... 其他字段
);
```

#### 支付重试历史表
```sql
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    attempt_number INT NOT NULL,
    attempted_at TIMESTAMP NOT NULL,
    status VARCHAR(20) NOT NULL,
    -- ... 其他字段
);
```

### 服务架构

```
┌─────────────────────────────────────────────────────────────┐
│                     支付模块                              │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryService                                        │
│  ├─── InitiateRetry()                                       │
│  ├─── ProcessPendingRetries()                               │
│  ├─── ClassifyFailure()                                     │
│  └─── GetRetryHealthMetrics()                               │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryWorker                                         │
│  ├─── handleProcessPaymentRetry()                           │
│  ├─── handleProcessPendingRetries()                         │
│  └─── handleRetryHealthCheck()                              │
├─────────────────────────────────────────────────────────────┤
│  PaymentRetryRepository                                     │
│  ├─── Create(), Update(), Delete()                          │
│  ├─── GetRetriesDueForProcessing()                          │
│  └─── GetRetryStatsByGateway()                              │
└─────────────────────────────────────────────────────────────┘
```

## 实现细节

### 1. 重试实体

#### PaymentRetry 实体
```go
type PaymentRetry struct {
    ID              uint      `json:"id"`
    PaymentRecordID uint      `json:"payment_record_id"`
    AttemptNumber   int       `json:"attempt_number"`
    MaxAttempts     int       `json:"max_attempts"`
    NextRetryAt     time.Time `json:"next_retry_at"`
    RetryStrategy   string    `json:"retry_strategy"`
    Status          string    `json:"status"`
    FailureType     string    `json:"failure_type"`
    // ... 其他字段
}
```

#### PaymentRetryHistory 实体
```go
type PaymentRetryHistory struct {
    ID              uint      `json:"id"`
    PaymentRetryID  uint      `json:"payment_retry_id"`
    AttemptNumber   int       `json:"attempt_number"`
    AttemptedAt     time.Time `json:"attempted_at"`
    Status          string    `json:"status"`
    Duration        int       `json:"duration"`
    ResponseCode    string    `json:"response_code"`
    // ... 其他字段
}
```

### 2. 重试策略

#### 默认重试配置
```go
var DefaultRetryStrategies = map[string]RetryStrategyConfig{
    PaymentGatewayEpay: {
        Gateway:           PaymentGatewayEpay,
        MaxAttempts:       3,
        InitialDelay:      3600,  // 1 小时
        MaxDelay:          86400, // 24 小时
        BackoffFactor:     2.0,
        Strategy:          RetryStrategyExponential,
        FailureTypes:      []string{FailureTypeTemporary, FailureTypeNetwork, FailureTypeGateway},
        TimeoutSeconds:    30,
        EnableAfterHours:  true,
        MaxConcurrent:     5,
    },
    // ... 其他网关
}
```

#### 指数退避算法
```go
func (pr *PaymentRetry) calculateExponentialDelay() int {
    if pr.AttemptNumber == 0 {
        return pr.InitialDelay
    }
    
    // 计算: initial_delay * (backoff_factor ^ attempt_number)
    delay := float64(pr.InitialDelay)
    for i := 0; i < pr.AttemptNumber; i++ {
        delay *= pr.BackoffFactor
    }
    
    // 确保不超过最大延迟
    if int(delay) > pr.MaxDelay {
        return pr.MaxDelay
    }
    
    return int(delay)
}
```

### 3. 故障分类

#### 自动故障类型检测
```go
func (s *paymentRetryService) ClassifyFailure(ctx context.Context, gateway, paymentMethod, errorCode, errorMessage string) string {
    // 永久性故障模式
    permanentPatterns := []string{
        "invalid card", "card expired", "insufficient funds",
        "card declined", "invalid account", "account closed",
    }
    
    // 网络故障模式
    networkPatterns := []string{
        "timeout", "network", "connection", "dns", "ssl",
    }
    
    // 网关故障模式
    gatewayPatterns := []string{
        "gateway", "service unavailable", "server error", "maintenance",
    }
    
    // 分类逻辑...
}
```

### 4. 任务队列集成

#### 重试任务工作器
```go
type PaymentRetryWorker struct {
    retryService interfaces.PaymentRetryService
    processor    *queue.TaskProcessor
}

func (w *PaymentRetryWorker) handleProcessPaymentRetry(ctx context.Context, task *asynq.Task) error {
    var payload struct {
        RetryID uint `json:"retry_id"`
    }
    
    // 解析负载并处理重试
    // ...
}
```

#### 计划任务处理
- **待处理重试**: 每 5 分钟处理一次
- **健康检查**: 每 15 分钟执行一次  
- **清理任务**: 每日凌晨 2 点进行维护
- **重试通知**: 重试事件的实时通知

## API 端点

### 管理员重试管理

#### 获取支付重试
```http
GET /admin/payments/retries
查询参数:
- user_id: 按用户 ID 过滤
- gateway: 按网关过滤 (epay, epusdt)
- status: 按重试状态过滤
- failure_type: 按故障类型过滤
- from_date, to_date: 日期范围过滤
- limit, offset: 分页
```

#### 获取重试详情
```http
GET /admin/payments/retries/{id}
响应: 包含完整历史记录的详细重试信息
```

#### 取消重试
```http
POST /admin/payments/retries/{id}/cancel
请求体: { "reason": "管理员手动取消" }
```

#### 重置重试
```http
POST /admin/payments/retries/{id}/reset
```

#### 批量操作
```http
POST /admin/payments/retries/bulk-cancel
POST /admin/payments/retries/bulk-reset
请求体: { "retry_ids": [1,2,3], "reason": "批量操作" }
```

#### 统计和健康状态
```http
GET /admin/payments/retries/statistics?gateway=epay&days=30
GET /admin/payments/retries/health
```

## 配置

### 环境变量
```bash
# 重试系统配置
RETRY_ENABLED=true
RETRY_MAX_CONCURRENT=100
RETRY_PROCESSING_INTERVAL=5m
RETRY_HEALTH_CHECK_INTERVAL=15m

# 通知设置
RETRY_NOTIFICATIONS_ENABLED=true
RETRY_NOTIFY_ON_FAILURE=true
RETRY_NOTIFY_ON_SUCCESS=false
```

### 网关特定配置
```go
// 运行时配置更新
retryService.UpdateRetryStrategy(ctx, "epay", "alipay", &RetryStrategyConfig{
    MaxAttempts:      5,
    InitialDelay:     1800, // 30 分钟
    MaxDelay:         43200, // 12 小时
    BackoffFactor:    1.5,
    Strategy:         RetryStrategyExponential,
})
```

## 监控和告警

### 健康指标
- **活跃重试总数**: 当前待处理/进行中的重试数量
- **逾期重试**: 应该已被处理但未处理的重试
- **成功率**: 24 小时和 7 天的成功率
- **网关健康状态**: 每个网关的性能指标

### 自动告警
- 待处理重试数量过多 (>50)
- 逾期重试数量过多 (>10)
- 成功率过低 (<80%)
- 网关特定健康问题

### 日志记录
所有重试操作都通过结构化日志进行全面记录:
```go
logger.Info("支付重试启动成功",
    logger.Uint("retry_id", retry.ID),
    logger.Uint("payment_record_id", paymentRecord.ID),
    logger.String("next_retry_at", retry.NextRetryAt.String()),
)
```

## 测试

### 单元测试
全面测试覆盖:
- 重试服务操作
- 指数退避计算
- 故障分类逻辑
- 仓储操作
- 工作器任务处理

### 集成测试
- 端到端重试工作流
- 数据库操作
- 队列集成
- API 端点测试

## 迁移

### 数据库迁移
```bash
# 应用重试表迁移
make migrate-up

# 验证迁移状态
make migrate-status
```

### 部署步骤
1. **部署数据库变更**: 应用迁移 000017
2. **部署应用程序**: 使用重试服务更新应用程序
3. **配置重试策略**: 设置网关特定配置
4. **启用处理**: 启动重试工作器和处理
5. **监控健康状态**: 验证系统健康状态和指标

## 性能考虑

### 可扩展性
- **水平扩展**: 多个工作器实例可以处理重试
- **队列分区**: 不同优先级的独立队列
- **数据库优化**: 重试查询的适当索引
- **缓存集成**: 缓存策略和配置

### 资源管理
- **内存使用**: 高效的数据结构和对象池
- **数据库连接**: 连接池和查询优化
- **队列处理**: 可配置的并发限制
- **清理**: 自动清理旧重试记录

## 安全

### 访问控制
- **仅管理员访问**: 重试管理仅限管理员用户
- **频率限制**: API 端点频率限制
- **输入验证**: 全面的请求验证
- **审计日志**: 记录所有管理员操作

### 数据保护
- **敏感数据**: 日志中的支付数据正确脱敏
- **加密**: 敏感配置数据加密
- **访问日志**: 全面的访问日志用于安全监控

## 故障排除

### 常见问题

#### 重试队列深度过高
**症状**: 大量待处理重试
**原因**: 网关问题、网络问题、配置错误
**解决方案**: 
- 检查网关健康状态
- 验证网络连接
- 检查重试配置
- 扩展工作器实例

#### 成功率过低
**症状**: 重试持续失败
**原因**: 永久性故障被重试、分类错误
**解决方案**:
- 检查故障分类规则
- 更新重试策略
- 检查网关特定问题

#### 性能问题
**症状**: 重试处理缓慢
**原因**: 数据库瓶颈、队列过载
**解决方案**:
- 优化数据库查询
- 扩展工作器实例
- 检查队列配置

## 未来增强

### 计划功能
1. **机器学习分类**: AI 驱动的故障分类
2. **自适应重试策略**: 基于成功率的动态策略调整
3. **高级分析**: 支付成功的预测分析
4. **多区域支持**: 分布式重试处理
5. **增强通知**: 带有支付上下文的丰富通知

### 性能优化
1. **缓存层**: 基于 Redis 的配置缓存
2. **数据库分片**: 水平数据库扩展
3. **异步处理**: 非阻塞重试处理
4. **批量操作**: 批量重试处理能力

## 结论

智能支付重试策略实现提供了一个强大、可扩展且智能的支付故障处理解决方案。通过全面的监控、灵活的配置和强大的管理工具，该系统显著提高了支付成功率，同时减少了运营开销。

该实现遵循清洁架构原则，提供广泛的测试覆盖，并包含用于维护和未来增强的全面文档。