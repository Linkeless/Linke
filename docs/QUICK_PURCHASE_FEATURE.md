# 快捷购买功能

## 概述

快捷购买功能允许用户通过单个 API 调用直接处理支付创建来购买订阅。这简化了订阅购买流程，相比传统的订单 → 发票 → 支付流程，减少了所需的步骤数量。

## 实现详情

### 架构

快捷购买功能遵循项目的清洁架构模式，并与现有的订阅、支付和优惠券领域集成。

### 关键组件

1. **处理器**: `/internal/domains/subscription/handlers/quick_purchase_handler.go`
   - 处理快捷购买的 HTTP 请求
   - 验证用户身份认证和输入
   - 提取客户端 IP 用于支付处理

2. **服务接口**: `/internal/domains/subscription/usecases/interfaces/subscription_order_service.go`
   - 在服务接口中添加了 `QuickPurchase` 方法
   - 定义了 `QuickPurchaseRequest` 和 `QuickPurchaseResponse` 结构

3. **服务实现**: `/internal/domains/subscription/usecases/implementations/subscription_order.go`
   - 实现核心快捷购买逻辑
   - 验证订阅计划可用性
   - 如果提供了优惠券，应用优惠券折扣
   - 直接创建支付，无需先创建订单/发票

4. **路由注册**: `/cmd/server/main.go`
   - 在 `POST /api/v1/subscription/quick-purchase` 注册快捷购买端点

## API 使用

### 端点
```
POST /api/v1/subscription/quick-purchase
```

### 身份认证
需要 Bearer token 认证。

### 请求体
```json
{
  "user_id": 1,
  "plan_id": 1,
  "payment_gateway": "epay",
  "payment_method": "alipay",
  "coupon_code": "SAVE20",
  "client_ip": "192.168.1.1",
  "return_url": "https://example.com/payment/return",
  "metadata": "{\"source\": \"web\"}"
}
```

### 响应
```json
{
  "status": "success",
  "message": "快捷购买创建成功",
  "data": {
    "payment_record": {
      "payment_no": "PAY20240101123456",
      "status": "pending",
      "amount": 29.99,
      "currency": "USD"
    },
    "payment_url": "https://payment.gateway.com/pay/123456",
    "qr_code_url": "https://payment.gateway.com/qr/123456",
    "expired_at": "2024-01-01T12:30:00Z",
    "plan_info": {
      "id": 1,
      "name": "高级套餐",
      "price": 29.99,
      "currency": "USD",
      "billing_cycle": "monthly"
    },
    "discount_info": {
      "coupon_code": "SAVE20",
      "discount_amount": 5.99,
      "original_amount": 29.99,
      "final_amount": 24.00
    }
  }
}
```

## 流程描述

### 快捷购买流程
1. **验证**: 
   - 检查重复的待处理订单
   - 验证订阅计划可用性
   - 如果提供了优惠券，进行验证

2. **价格计算**:
   - 计算基础金额和设置费用
   - 应用优惠券折扣
   - 验证最终金额是否合理

3. **支付创建**:
   - 直接创建支付订单
   - 向用户返回支付 URL
   - 此阶段不创建订单或发票

4. **异步处理**:
   - 支付成功后，自动创建订单和发票
   - 支付成功后激活订阅

### 安全功能

- **重复订单预防**: 检查现有的待处理订单
- **频率限制**: 通过 5 分钟冷却时间防止垃圾请求
- **价格保护**: 验证金额并防止过度折扣
- **输入验证**: 验证所有输入参数
- **身份认证**: 需要有效的用户身份认证

### 错误处理

API 返回适当的 HTTP 状态码和错误消息：

- `400 Bad Request`: 无效的输入数据或验证错误
- `401 Unauthorized`: 缺少或无效的身份认证
- `403 Forbidden`: 用户尝试为其他用户创建购买
- `500 Internal Server Error`: 服务器端错误

## 测试

实现包含以下位置的综合单元测试：
- `/internal/domains/subscription/handlers/quick_purchase_handler_test.go`

测试覆盖：
- 成功的快捷购买场景
- 身份认证失败
- 无效的请求数据
- 服务层错误

## 优势

1. **改善用户体验**: 单个 API 调用即可完成订阅购买
2. **降低延迟**: 减少客户端和服务器之间的往返次数
3. **简化集成**: 前端应用程序更容易实现
4. **维护安全性**: 保留所有现有的安全验证
5. **向后兼容**: 现有的订单/发票流程仍然可用

## 未来增强

1. **异步订单创建**: 为支付成功后的订单/发票创建实现适当的异步处理
2. **支付 Webhook**: 处理支付状态更新以触发订单/订阅创建
3. **重试机制**: 为失败的异步操作添加重试逻辑
4. **指标**: 为快捷购买成功率添加监控和指标

## 配置

无需额外配置。该功能使用现有的：
- 支付网关配置
- 优惠券服务设置
- 订阅计划数据
- 安全设置