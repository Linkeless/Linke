# 事件处理器测试指南

本指南为新实现的跨领域事件处理系统提供全面的测试说明。

## 概述

事件处理系统已经完全重构，通过事件驱动架构支持完整的跨领域业务流程。关键功能包括：

- **完整的业务流程集成**: 支付 → 订单处理 → 订阅激活 → 发票生成 → 使用量监控
- **幂等性保护**: 所有事件处理器都包含重复处理预防机制
- **错误恢复**: 具有部分故障容错的强健错误处理
- **实时使用量监控**: 自动流量限制检查和警报
- **异步处理**: 邮件通知和繁重任务异步运行

## 测试策略

### 1. 构建和依赖验证

首先，验证所有依赖项是否正确解析：

```bash
# 构建应用程序检查编译错误
make build

# 检查导入问题
go mod tidy
go mod verify
```

### 2. 单元测试事件处理器

独立测试各个事件处理器：

```bash
# 测试事件系统组件
go test ./internal/shared/events/... -v

# 测试特定处理器（如果存在单独测试）
go test ./internal/shared/events/ -run TestPaymentCompletedHandler -v
go test ./internal/shared/events/ -run TestSubscriptionCreatedHandler -v
```

### 3. 集成测试

#### 3.1 完整支付流程测试

测试完整的支付 → 订阅激活流程：

```bash
# 启动应用程序
make safe-dev

# 在另一个终端中测试流程
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "payment.completed",
    "payment_id": "test_payment_123",
    "user_id": 1,
    "amount": 29.99,
    "order_id": 1,
    "currency": "USD"
  }'
```

**预期流程：**
1. PaymentCompletedHandler 处理支付
2. 订单状态更新为 "paid"  
3. 创建/续费订阅
4. 生成发票
5. 欢迎/确认邮件加入队列
6. 用户状态激活

#### 3.2 用户注册流程测试

测试用户注册流程：

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "username": "testuser",
    "password": "securepassword123",
    "first_name": "Test",
    "last_name": "User"
  }'
```

**预期流程：**
1. 用户在数据库中创建
2. UserRegisteredHandler 触发
3. 欢迎邮件加入队列
4. 用户配置缓存初始化
5. 账户准备好进行订阅

#### 3.3 流量监控测试

测试使用量监控和限制执行：

```bash
# 模拟订阅的流量使用
curl -X POST http://localhost:8080/api/v1/subscriptions/1/usage \
  -H "Content-Type: application/json" \
  -d '{
    "bytes_used": 85000000000,
    "subscription_id": 1
  }'
```

**预期流程：**
1. 数据库中使用量更新
2. 在80%、90%时触发警告事件
3. 在100%时触发限制超出事件
4. 超出限制时账户被暂停
5. 向用户发送警报邮件

### 4. 事件系统监控

#### 4.1 事件总线健康检查

```bash
# 检查事件系统状态
curl http://localhost:8080/health

# 监控应用日志中的事件处理
tail -f logs/application.log | grep -i "event"
```

#### 4.2 幂等性测试

测试重复事件是否被正确处理：

```bash
# 发送相同的支付完成事件两次
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "duplicate_test_123",
    "event_type": "payment.completed", 
    "payment_id": "test_payment_456",
    "user_id": 1,
    "amount": 49.99,
    "order_id": 2
  }'

# 立即再次发送 - 应该被跳过
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "duplicate_test_123",
    "event_type": "payment.completed",
    "payment_id": "test_payment_456", 
    "user_id": 1,
    "amount": 49.99,
    "order_id": 2
  }'
```

**预期结果：** 第二个请求应该记录 "already processed, skipping"

### 5. 数据库状态验证

运行测试后，验证数据库状态：

```sql
-- 检查事件处理记录
SELECT * FROM events ORDER BY created_at DESC LIMIT 10;

-- 检查订阅状态  
SELECT id, user_id, status, traffic_used, traffic_limit, traffic_suspended 
FROM user_subscriptions WHERE user_id = 1;

-- 检查订单处理
SELECT id, user_id, status, total_amount, created_at 
FROM subscription_orders ORDER BY created_at DESC LIMIT 5;

-- 检查发票生成
SELECT id, user_id, status, amount, invoice_number 
FROM invoices ORDER BY created_at DESC LIMIT 5;
```

### 6. 异步任务验证

检查异步任务是否正在被处理：

```bash
# 监控任务队列处理
redis-cli MONITOR | grep -i "task\|queue\|email\|notification"

# 检查任务队列状态
redis-cli
> LLEN queue:email
> LLEN queue:notification  
> LLEN queue:data_processing
```

### 7. 缓存系统测试

验证缓存行为：

```bash
# 检查缓存键
redis-cli KEYS "event_processed:*"
redis-cli KEYS "user_config:*" 
redis-cli KEYS "user_billing:*"

# 验证缓存TTL
redis-cli TTL "event_processed:some_event_id"
```

## 负载测试

### 事件处理压力测试

```bash
# 安装Apache Bench（如果没有）
brew install httpd  # macOS
# 或
apt-get install apache2-utils  # Ubuntu

# 测试支付webhook处理
ab -n 100 -c 10 -T application/json -p payment_data.json \
   http://localhost:8080/api/v1/payments/webhook

# payment_data.json 内容：
{
  "event_type": "payment.completed",
  "payment_id": "load_test_{{#}}",
  "user_id": 1,
  "amount": 29.99,
  "order_id": 1
}
```

## 错误场景测试

### 1. 服务依赖故障

测试服务不可用时的行为：

```bash
# 模拟数据库连接问题
# 临时停止数据库并发送事件

# 模拟Redis缓存不可用  
# 停止Redis并测试事件处理

# 使用无效数据测试
curl -X POST http://localhost:8080/api/v1/payments/webhook \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "payment.completed",
    "payment_id": "",
    "user_id": "invalid",
    "amount": -10,
    "order_id": "not_number"
  }'
```

### 2. 部分处理失败

测试优雅降级：

- 邮件服务停机 → 事件仍然处理，邮件排队重试
- 发票服务错误 → 支付仍然处理，发票创建重试
- 缓存不可用 → 事件处理不使用缓存（较慢但功能正常）

## 性能监控

### 要跟踪的关键指标

1. **事件处理延迟**
   - 从事件发布到完成的平均时间
   - 目标：简单事件 < 100ms，复杂流程 < 500ms

2. **吞吐量**  
   - 每秒处理的事件数
   - 目标：> 100 事件/秒

3. **错误率**
   - 事件处理失败百分比  
   - 目标：< 1% 失败率

4. **队列深度**
   - 异步任务队列长度
   - 目标：正常负载下队列在30秒内清空

### 监控命令

```bash
# 应用程序指标
curl http://localhost:8080/metrics

# 事件处理统计
grep "event processed" logs/application.log | wc -l

# 错误跟踪
grep "ERROR" logs/application.log | tail -20
```

## 常见问题故障排除

### 问题：事件未处理
**诊断：**
```bash
# 检查事件总线初始化
grep "Event system initialized" logs/application.log

# 验证处理器注册
grep "event handlers registered" logs/application.log
```

### 问题：内存使用率高
**诊断：**
```bash
# 检查已处理事件缓存
redis-cli DBSIZE
redis-cli KEYS "event_processed:*" | wc -l
```

### 问题：重复处理
**诊断：**  
```bash
# 检查幂等性缓存命中
grep "already processed, skipping" logs/application.log
```

## 回滚计划

如果生产环境出现问题：

1. **立即回滚：** 恢复到之前的 handlers.go 版本
2. **部分禁用：** 注释掉特定的处理器注册  
3. **紧急模式：** 完全禁用事件系统（手动处理）

```go
// 在 bootstrap/app.go 中紧急禁用
// 注释掉事件处理器注册：
/*
if err := crossDomainHandlers.RegisterCrossDomainHandlers(eventBus); err != nil {
    logger.Error("Failed to register cross-domain event handlers", zap.Error(err))
}
*/
```

## 成功标准

实现成功的标准：

- ✅ 解决所有编译错误
- ✅ 支付 → 订阅流程端到端工作  
- ✅ 用户注册触发欢迎流程
- ✅ 流量监控正确生成警报
- ✅ 幂等性防止重复处理
- ✅ 错误处理维护系统稳定性  
- ✅ 异步任务在可接受的时间内处理
- ✅ 数据库状态在故障中保持一致
- ✅ 性能满足定义的目标

## 下一步计划

成功测试后：

1. **生产部署：** 使用功能标志部署
2. **监控设置：** 配置错误率警报
3. **文档更新：** 更新API文档
4. **团队培训：** 向团队介绍新的事件流程
5. **逐步推出：** 逐步启用功能