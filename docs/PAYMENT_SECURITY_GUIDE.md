# 支付回调安全性增强指南

本文档提供了支付回调处理的安全性改进方案和最佳实践建议。

## 概述

支付回调是金融交易中最关键的环节之一，需要实施多层安全防护措施来确保交易的安全性和完整性。本项目实现了以下安全机制：

## 已实现的安全机制

### 1. 签名验证 (Signature Verification)

**实现位置**: `internal/shared/middleware/payment_security.go`

**功能**:
- 支持多种签名算法（MD5, HMAC-SHA256）
- 针对不同支付网关使用相应的签名验证方式
- Epay: MD5签名验证
- EPUSDT: HMAC-SHA256签名验证

**配置示例**:
```env
PAYMENT_REQUIRE_SIGNATURE=true
EPAY_SIGN_KEY=your_epay_signature_key
EPUSDT_SIGN_KEY=your_epusdt_signature_key
```

### 2. 防重放攻击 (Replay Attack Prevention)

**实现位置**: `internal/shared/middleware/payment_security.go`, `internal/domains/payment/usecases/implementations/payment.go`

**功能**:
- Redis基础的请求去重
- 时间窗口验证（默认5分钟）
- 请求内容哈希验证
- 通知间隔检查（最小30秒间隔）

**配置示例**:
```env
PAYMENT_ENABLE_REPLAY_PROTECTION=true
PAYMENT_REPLAY_TIME_WINDOW_MINUTES=5
```

### 3. IP白名单验证 (IP Whitelist)

**实现位置**: `internal/shared/middleware/payment_security.go`

**功能**:
- 支持单个IP地址验证
- 支持CIDR网段验证
- 分网关配置不同的IP白名单

**配置示例**:
```env
PAYMENT_ENABLE_IP_WHITELIST=true
EPAY_IP_WHITELIST=192.168.1.100,10.0.0.0/8
EPUSDT_IP_WHITELIST=203.107.32.1,203.107.33.1
```

### 4. 请求频率限制 (Rate Limiting)

**实现位置**: `internal/shared/middleware/rate_limit.go`

**功能**:
- 基于令牌桶算法的限流
- 按IP+网关组合进行限制
- 可配置的限流参数

**配置示例**:
```env
PAYMENT_NOTIFY_RATE_LIMIT=10
PAYMENT_NOTIFY_RATE_BURST=2
```

### 5. 敏感信息保护 (Sensitive Data Protection)

**实现位置**: `internal/domains/payment/entities/payment_record.go`

**功能**:
- 交易ID部分掩码显示
- 过期支付链接自动清理
- 分级响应数据结构
- 内部字段过滤

### 6. 增强幂等性保证 (Enhanced Idempotency)

**实现位置**: `internal/domains/payment/usecases/implementations/payment.go`

**功能**:
- 内容哈希验证
- 时间间隔检查
- 来源IP监控
- 状态降级保护

## 安全配置参数详解

### 环境变量配置

```env
# 签名验证
PAYMENT_REQUIRE_SIGNATURE=true
EPAY_SIGN_KEY=your_32_character_sign_key_here
EPUSDT_SIGN_KEY=your_64_character_hmac_key_here

# IP白名单 (可选)
PAYMENT_ENABLE_IP_WHITELIST=false
EPAY_IP_WHITELIST=192.168.1.0/24,10.0.0.100
EPUSDT_IP_WHITELIST=203.107.32.0/24

# 重放攻击防护
PAYMENT_ENABLE_REPLAY_PROTECTION=true
PAYMENT_REPLAY_TIME_WINDOW_MINUTES=5

# 速率限制
PAYMENT_NOTIFY_RATE_LIMIT=10
PAYMENT_NOTIFY_RATE_BURST=2

# 其他安全选项
PAYMENT_REQUIRE_HTTPS=true
PAYMENT_MAX_REQUEST_SIZE=1048576
```

## 部署安全建议

### 1. 网络层安全

```bash
# 配置防火墙规则
sudo ufw allow from 203.107.32.0/24 to any port 8080
sudo ufw allow from 203.107.33.0/24 to any port 8080

# 配置Nginx反向代理
server {
    listen 443 ssl;
    server_name yourdomain.com;
    
    location /api/v1/payment/notify/ {
        # 限制请求体大小
        client_max_body_size 1M;
        
        # 设置超时
        proxy_read_timeout 30s;
        proxy_send_timeout 30s;
        
        # IP白名单 (如果不使用应用层白名单)
        allow 203.107.32.0/24;
        allow 203.107.33.0/24;
        deny all;
        
        proxy_pass http://localhost:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 2. 数据库安全

```sql
-- 创建支付回调专用数据库用户
CREATE USER 'payment_callback'@'localhost' IDENTIFIED BY 'strong_password';
GRANT SELECT, UPDATE ON linke.payment_records TO 'payment_callback'@'localhost';
GRANT SELECT ON linke.payment_configs TO 'payment_callback'@'localhost';

-- 定期清理过期的重放保护记录
DELETE FROM payment_records 
WHERE last_notify_time < DATE_SUB(NOW(), INTERVAL 7 DAY) 
AND status IN ('failed', 'cancelled');
```

### 3. Redis安全配置

```redis
# redis.conf 安全配置
bind 127.0.0.1
requirepass your_redis_password
maxmemory 256mb
maxmemory-policy allkeys-lru

# 配置专用的支付回调键空间
config set notify-keyspace-events KEA
```

## 监控和告警

### 1. 安全事件监控

```go
// 关键安全事件日志格式
{
  "timestamp": "2024-01-01T10:00:00Z",
  "level": "WARN",
  "event_type": "SIGNATURE_VALIDATION_FAILED",
  "gateway": "epay",
  "client_ip": "1.2.3.4",
  "user_agent": "curl/7.68.0",
  "payment_no": "PAY20240101001"
}
```

### 2. 推荐的监控指标

- 签名验证失败率
- IP白名单违规次数
- 重放攻击检测次数
- 通知处理延迟
- 限流触发频率

### 3. 告警阈值建议

```yaml
alerts:
  - name: payment_security_signature_failures
    condition: rate(signature_validation_failed[5m]) > 10
    severity: critical
    
  - name: payment_security_ip_violations
    condition: rate(ip_whitelist_violation[5m]) > 5
    severity: warning
    
  - name: payment_security_replay_attacks
    condition: rate(replay_attack_detected[5m]) > 3
    severity: critical
```

## 测试安全性

### 1. 运行安全测试

```bash
# 运行支付安全测试套件
go test -v ./internal/domains/payment/handlers/ -run TestPaymentNotify_Security

# 运行性能基准测试
go test -bench=BenchmarkPaymentNotify ./internal/domains/payment/handlers/

# 运行完整的安全测试
make security-test
```

### 2. 手动安全测试

```bash
# 测试签名验证
curl -X POST http://localhost:8080/api/v1/payment/notify/epay \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "out_trade_no=test123&trade_status=TRADE_SUCCESS&sign=invalid"

# 测试IP白名单
curl -X POST http://localhost:8080/api/v1/payment/notify/epay \
  -H "X-Real-IP: 1.2.3.4" \
  -H "Content-Type: application/json" \
  -d '{"out_trade_no":"test123","trade_status":"TRADE_SUCCESS"}'

# 测试请求大小限制
curl -X POST http://localhost:8080/api/v1/payment/notify/epay \
  -H "Content-Type: application/json" \
  -d "$(python -c 'print("{\"data\":\"" + "x"*2000000 + "\"}")')"
```

## 安全最佳实践

### 1. 密钥管理

- 使用强随机密钥生成器
- 定期轮换签名密钥
- 使用密钥管理服务 (如 HashiCorp Vault)
- 避免在代码中硬编码密钥

### 2. 错误处理

- 不要在错误响应中泄露内部信息
- 统一错误响应格式
- 记录详细的内部错误日志
- 实施错误响应的延迟机制

### 3. 日志记录

- 记录所有安全相关事件
- 使用结构化日志格式
- 实施日志轮转和归档
- 保护日志文件访问权限

### 4. 定期安全审计

- 每季度检查IP白名单
- 定期更新签名密钥
- 审查异常访问模式
- 更新安全配置参数

## 紧急响应程序

### 1. 安全事件响应

```bash
# 临时禁用特定IP
redis-cli SET "payment_blocked_ip:1.2.3.4" "1" EX 3600

# 临时禁用特定网关
redis-cli SET "payment_gateway_disabled:epay" "1" EX 1800

# 启用紧急模式 (仅处理已知好IP)
export PAYMENT_EMERGENCY_MODE=true
export PAYMENT_EMERGENCY_IPS="192.168.1.100,10.0.0.0/8"
```

### 2. 回滚程序

```bash
# 回滚到安全的配置版本
git checkout v1.2.3 -- internal/shared/config/
make deploy-emergency

# 禁用所有支付回调
export PAYMENT_MAINTENANCE_MODE=true
```

## 性能考虑

### 1. Redis优化

- 使用连接池减少连接开销
- 设置合适的过期时间
- 监控Redis内存使用

### 2. 中间件优化

- 按重要性排序中间件
- 使用异步日志记录
- 实施熔断机制

### 3. 数据库优化

- 为安全字段添加索引
- 定期清理历史数据
- 使用读写分离

## 结论

通过实施这些安全措施，支付回调系统能够有效防御常见的安全攻击，确保交易数据的完整性和系统的稳定性。建议在生产环境部署前进行充分的安全测试，并持续监控系统的安全状态。