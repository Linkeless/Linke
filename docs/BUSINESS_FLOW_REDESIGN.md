# 标准商务流程重构设计

## 一、业务流程重新定义

### 当前问题
- 订单 → 付款 → 发票 (错误流程)
- Invoice和Order功能重复
- 缺乏标准商务合规性

### 标准流程
```
1. 用户选择服务 → 2. 创建订单 → 3. 生成发票 → 4. 用户付款 → 5. 确认收款 → 6. 激活服务
```

## 二、新数据模型设计

### 核心实体关系
```
User → Order → Invoice → Payment → Subscription
  ↓       ↓        ↓        ↓         ↓
订阅者   购买意向   付款请求   资金流转   服务履行
```

### 状态流转设计

#### Order状态流转
```
draft → confirmed → invoiced → paid → fulfilled → cancelled
草稿     确认        已开票     已付款    已履行      已取消
```

#### Invoice状态流转  
```
draft → sent → paid → overdue → cancelled → voided
草稿    已发送  已付款   逾期      已取消      已作废
```

#### Payment状态流转
```
pending → processing → completed → failed → refunded
待处理     处理中        完成       失败      已退款
```

## 三、数据表结构重新设计

### 1. Orders表调整
```sql
-- 订单专注于购买意向和服务配置
CREATE TABLE subscription_orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_number VARCHAR(32) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    subscription_plan_id BIGINT NOT NULL,
    
    -- 订单基本信息
    status ENUM('draft', 'confirmed', 'invoiced', 'paid', 'fulfilled', 'cancelled') DEFAULT 'draft',
    order_type ENUM('new', 'renewal', 'upgrade', 'downgrade') DEFAULT 'new',
    
    -- 服务配置
    billing_cycle ENUM('monthly', 'quarterly', 'annually') NOT NULL,
    billing_interval INT DEFAULT 1,
    service_start_date TIMESTAMP NULL,
    service_end_date TIMESTAMP NULL,
    
    -- 定价信息
    base_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    setup_fee DECIMAL(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- 优惠券信息
    coupon_code VARCHAR(50) NULL,
    coupon_discount_amount DECIMAL(10,2) DEFAULT 0,
    
    -- 业务字段
    notes TEXT NULL,
    metadata JSON NULL,
    
    -- 时间戳
    confirmed_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_order_number (order_number)
);
```

### 2. Invoices表调整
```sql
-- 发票专注于财务和合规
CREATE TABLE invoices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    invoice_number VARCHAR(32) UNIQUE NOT NULL,
    order_id BIGINT NOT NULL, -- 基于哪个订单
    user_id BIGINT NOT NULL,
    
    -- 发票状态
    status ENUM('draft', 'sent', 'paid', 'overdue', 'cancelled', 'voided') DEFAULT 'draft',
    invoice_type ENUM('standard', 'proforma', 'credit_note') DEFAULT 'standard',
    
    -- 金额信息
    subtotal DECIMAL(10,2) NOT NULL,
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- 账期信息
    issued_at TIMESTAMP NULL,
    due_at TIMESTAMP NULL,
    payment_terms_days INT DEFAULT 30,
    
    -- 收款信息
    paid_amount DECIMAL(10,2) DEFAULT 0,
    paid_at TIMESTAMP NULL,
    
    -- 开票信息
    billing_name VARCHAR(255) NOT NULL,
    billing_email VARCHAR(255) NOT NULL,
    billing_address TEXT NULL,
    billing_city VARCHAR(100) NULL,
    billing_state VARCHAR(100) NULL,
    billing_country VARCHAR(2) NULL,
    billing_zip VARCHAR(20) NULL,
    
    -- 税务信息
    tax_number VARCHAR(50) NULL,
    tax_type VARCHAR(20) NULL, -- VAT, GST, etc.
    
    -- 企业信息
    company_name VARCHAR(255) NULL,
    company_address TEXT NULL,
    company_tax_id VARCHAR(50) NULL,
    
    -- 发票内容
    description TEXT NOT NULL,
    line_items JSON NULL, -- 发票明细
    
    -- 文档管理
    pdf_path VARCHAR(500) NULL,
    pdf_size INT NULL,
    template VARCHAR(50) DEFAULT 'default',
    language VARCHAR(5) DEFAULT 'en',
    
    -- 发送记录
    sent_at TIMESTAMP NULL,
    send_count INT DEFAULT 0,
    last_reminder_at TIMESTAMP NULL,
    
    -- 作废信息
    voided_at TIMESTAMP NULL,
    void_reason TEXT NULL,
    
    -- 业务字段
    notes TEXT NULL,
    internal_notes TEXT NULL,
    metadata JSON NULL,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (order_id) REFERENCES subscription_orders(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_invoice_number (invoice_number),
    INDEX idx_due_at (due_at)
);
```

### 3. Payments表调整
```sql
-- 支付专注于资金流转
CREATE TABLE payments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_number VARCHAR(32) UNIQUE NOT NULL,
    invoice_id BIGINT NOT NULL, -- 支付哪张发票
    user_id BIGINT NOT NULL,
    
    -- 支付状态
    status ENUM('pending', 'processing', 'completed', 'failed', 'refunded', 'cancelled') DEFAULT 'pending',
    payment_intent_id VARCHAR(255) NULL, -- 第三方支付意向ID
    
    -- 金额信息
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    
    -- 支付渠道信息
    payment_method VARCHAR(50) NOT NULL, -- alipay, wechat, stripe, etc.
    payment_gateway VARCHAR(50) NOT NULL, -- epay, stripe, etc.
    gateway_transaction_id VARCHAR(255) NULL,
    gateway_fee DECIMAL(10,2) DEFAULT 0,
    
    -- 支付详情
    payment_url TEXT NULL,
    qr_code_url TEXT NULL,
    redirect_url TEXT NULL,
    
    -- 时间信息
    expires_at TIMESTAMP NULL,
    processed_at TIMESTAMP NULL,
    completed_at TIMESTAMP NULL,
    
    -- 退款信息
    refund_amount DECIMAL(10,2) DEFAULT 0,
    refunded_at TIMESTAMP NULL,
    refund_reason TEXT NULL,
    refund_reference VARCHAR(255) NULL,
    
    -- 通知信息
    webhook_data JSON NULL,
    notification_count INT DEFAULT 0,
    last_notification_at TIMESTAMP NULL,
    
    -- 业务字段
    notes TEXT NULL,
    metadata JSON NULL,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    FOREIGN KEY (invoice_id) REFERENCES invoices(id),
    FOREIGN KEY (user_id) REFERENCES users(id),
    INDEX idx_invoice_id (invoice_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_payment_method (payment_method),
    INDEX idx_gateway_transaction_id (gateway_transaction_id)
);
```

## 四、API重新设计

### 用户端API流程
```
1. POST /api/v1/orders              - 创建订单
2. POST /api/v1/orders/{id}/confirm - 确认订单
3. GET  /api/v1/orders/{id}/invoice - 获取发票
4. POST /api/v1/invoices/{id}/pay   - 发起付款
5. GET  /api/v1/payments/{id}       - 查询支付状态
```

### 管理端API流程
```
1. GET  /api/v1/admin/orders        - 订单管理
2. POST /api/v1/admin/invoices      - 手动开票
3. POST /api/v1/admin/invoices/{id}/send - 发送发票
4. POST /api/v1/admin/payments/{id}/confirm - 确认收款
```

## 五、业务逻辑重构

### 1. 订单服务 (OrderService)
```go
// 创建订单 - 只记录购买意向
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*Order, error)

// 确认订单 - 触发开票
func (s *OrderService) ConfirmOrder(orderID uint) (*Order, error)

// 履行订单 - 激活服务
func (s *OrderService) FulfillOrder(orderID uint) error
```

### 2. 发票服务 (InvoiceService)  
```go
// 基于订单生成发票
func (s *InvoiceService) GenerateFromOrder(orderID uint) (*Invoice, error)

// 发送发票给客户
func (s *InvoiceService) SendInvoice(invoiceID uint) error

// 标记发票为已付款
func (s *InvoiceService) MarkAsPaid(invoiceID uint, paymentID uint) error
```

### 3. 支付服务 (PaymentService)
```go
// 基于发票创建付款
func (s *PaymentService) CreatePayment(invoiceID uint, method string) (*Payment, error)

// 处理支付回调
func (s *PaymentService) HandleCallback(paymentID uint, data PaymentCallbackData) error

// 确认支付成功
func (s *PaymentService) ConfirmPayment(paymentID uint) error
```

## 六、实施策略

### 阶段1：数据模型迁移
1. 创建新表结构
2. 数据迁移脚本
3. 保持双写一段时间

### 阶段2：API逐步切换  
1. 新API并行开发
2. 前端逐步切换
3. 旧API保持兼容

### 阶段3：业务逻辑优化
1. 完善各服务间协作
2. 优化性能和错误处理
3. 完善监控和日志

### 阶段4：清理和优化
1. 移除旧代码和API
2. 数据清理
3. 性能优化

## 七、风险控制

### 技术风险
- 数据迁移风险：分批迁移，验证数据完整性
- API兼容性：保持向后兼容一段时间
- 性能风险：新流程的性能测试

### 业务风险  
- 用户体验：确保新流程不影响用户操作
- 财务风险：确保所有金额计算准确
- 合规风险：新发票流程符合财务规范

## 八、预期收益

### 业务收益
- 标准化的商务流程
- 更好的财务合规性
- 清晰的业务边界

### 技术收益
- 减少重复代码
- 更清晰的数据模型
- 更好的可维护性

### 运营收益
- 更好的财务管理
- 更清晰的业务报表
- 更完善的审计追踪