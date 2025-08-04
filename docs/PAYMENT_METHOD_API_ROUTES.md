# 支付方式管理 API 路由

本文档概述了在 Linke 平台中实现的自服务支付方式管理功能的 API 路由。

## 集成点

### 主服务器配置
支付方式处理器应添加到 `/cmd/server/main.go` 中的 `NewHTTPServer` 函数：

```go
func NewHTTPServer(
    // ... 现有参数
    paymentHandler *paymentHandlers.PaymentHandler,
    paymentMethodHandler *paymentHandlers.PaymentMethodHandler, // 添加这个
    // ... 其他参数
) *HTTPServer {
```

### API 路由
以下路由应添加到主服务器的支付组中：

```go
// 支付方法路由 (/api/v1/payment-methods)
paymentMethodGroup := apiV1.Group("/payment-methods")
{
    // 列出用户支付方式
    paymentMethodGroup.GET("", paymentMethodHandler.ListPaymentMethods)
    
    // 获取默认支付方式
    paymentMethodGroup.GET("/default", paymentMethodHandler.GetDefaultPaymentMethod)
    
    // 创建新支付方式
    paymentMethodGroup.POST("", paymentMethodHandler.CreatePaymentMethod)
    
    // 获取特定支付方式
    paymentMethodGroup.GET("/:id", paymentMethodHandler.GetPaymentMethod)
    
    // 更新支付方式
    paymentMethodGroup.PUT("/:id", paymentMethodHandler.UpdatePaymentMethod)
    
    // 删除支付方式
    paymentMethodGroup.DELETE("/:id", paymentMethodHandler.DeletePaymentMethod)
    
    // 设置支付方式为默认
    paymentMethodGroup.PUT("/:id/default", paymentMethodHandler.SetDefaultPaymentMethod)
    
    // 验证支付方式
    paymentMethodGroup.POST("/:id/validate", paymentMethodHandler.ValidatePaymentMethod)
    
    // 获取支付方式使用统计
    paymentMethodGroup.GET("/:id/stats", paymentMethodHandler.GetPaymentMethodUsageStats)
}
```

## API 端点文档

### 1. 列出支付方式
- **端点**: `GET /api/v1/payment-methods`
- **查询参数**:
  - `gateway` (字符串, 可选): 按支付网关过滤
  - `active_only` (布尔值, 可选): 仅显示活跃支付方式
- **响应**: 用户支付方式列表及默认方式信息

### 2. 获取默认支付方式
- **端点**: `GET /api/v1/payment-methods/default`
- **查询参数**:
  - `gateway` (字符串, 可选): 按支付网关过滤
- **响应**: 用户的默认支付方式

### 3. 创建支付方式
- **端点**: `POST /api/v1/payment-methods`
- **安全**: 安全限流
- **请求体**: 支付方式创建数据（已令牌化）
- **响应**: 已创建的支付方式信息

### 4. 获取支付方式
- **端点**: `GET /api/v1/payment-methods/{id}`
- **安全**: 用户所有权验证
- **响应**: 支付方式详情

### 5. 更新支付方式
- **端点**: `PUT /api/v1/payment-methods/{id}`
- **安全**: 用户所有权验证
- **请求体**: 支付方式更新数据
- **响应**: 已更新的支付方式信息

### 6. 删除支付方式
- **端点**: `DELETE /api/v1/payment-methods/{id}`
- **安全**: 用户所有权验证
- **响应**: 成功确认

### 7. 设置默认支付方式
- **端点**: `PUT /api/v1/payment-methods/{id}/default`
- **安全**: 用户所有权验证
- **响应**: 已更新的支付方式及默认状态

### 8. 验证支付方式
- **端点**: `POST /api/v1/payment-methods/{id}/validate`
- **安全**: 限流，用户所有权验证
- **响应**: 验证结果和已更新的支付方式

### 9. 获取使用统计
- **端点**: `GET /api/v1/payment-methods/{id}/stats`
- **安全**: 用户所有权验证
- **响应**: 支付方式的使用统计

## 已实现的安全功能

1. **PCI 合规**: 不存储原始支付数据，仅存储令牌化数据
2. **用户所有权验证**: 所有操作验证用户所有权
3. **限流**: 敏感操作（创建、验证）实施限流
4. **审计日志**: 所有操作都记录用于安全跟踪
5. **错误掩码**: API 响应中屏蔽敏感错误

## 与订阅订单集成

订阅订单创建已增强以支持：

1. **默认支付方式**: 使用 `use_default_payment: true`
2. **特定支付方式**: 使用 `payment_method_id: <id>`
3. **传统方式**: 指定 `payment_gateway` 和 `payment_method`

使用已保存支付方式的订阅订单请求示例：
```json
{
  "subscription_plan_id": 1,
  "order_type": "new",
  "use_default_payment": true,
  "coupon_code": "SAVE20"
}
```

## 数据库迁移

运行迁移以创建 payment_methods 表：
```bash
make migrate-up
```

迁移文件 `000018_create_payment_methods_table.up.sql` 包括：
- 安全导向的表结构设计
- 性能优化的适当索引
- 确保数据完整性的约束
- 自动字段更新的触发器
- 执行业务规则的函数

## 测试

已为以下内容实现基础测试：
- 支付方式实体验证
- 服务常量和配置
- 请求/响应结构

进行全面测试，请运行：
```bash
make test
```

## 未来增强

1. **高级分析**: 支付方式性能分析
2. **智能推荐**: AI 驱动的支付方式建议
3. **欺诈检测**: 支付方式的高级欺诈检测
4. **国际支持**: 多货币和区域支付方式
5. **订阅优化**: 使用不同支付方式自动重试