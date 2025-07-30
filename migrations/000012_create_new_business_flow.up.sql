-- 全新商务流程数据库结构迁移 - 简洁版
-- 业务逻辑由代码层处理，数据库只负责数据存储

-- 1. 创建订单表 - 购买意向和服务配置
CREATE TABLE orders (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    order_number VARCHAR(32) UNIQUE NOT NULL,
    user_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    
    -- 订单状态
    status VARCHAR(20) DEFAULT 'draft',
    order_type VARCHAR(20) DEFAULT 'new',
    
    -- 服务配置
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT DEFAULT 1,
    service_period INT NOT NULL COMMENT '服务期长度（月）',
    
    -- 定价信息（订单创建时锁定价格）
    base_amount DECIMAL(10,2) NOT NULL,
    discount_amount DECIMAL(10,2) DEFAULT 0,
    setup_fee DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- 优惠券信息
    coupon_code VARCHAR(50) NULL,
    coupon_discount DECIMAL(10,2) DEFAULT 0,
    
    -- 服务时间
    service_start_date TIMESTAMP NULL,
    service_end_date TIMESTAMP NULL,
    
    -- 状态时间戳
    confirmed_at TIMESTAMP NULL,
    cancelled_at TIMESTAMP NULL,
    fulfilled_at TIMESTAMP NULL,
    
    -- 业务字段
    notes TEXT NULL,
    metadata JSON NULL,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- 索引
    INDEX idx_user_id (user_id),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status),
    INDEX idx_order_number (order_number),
    INDEX idx_created_at (created_at),
    INDEX idx_user_status_created (user_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订单表';

-- 2. 创建发票表 - 付款请求和财务合规
CREATE TABLE invoices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    invoice_number VARCHAR(32) UNIQUE NOT NULL,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    
    -- 发票状态
    status VARCHAR(20) DEFAULT 'draft',
    invoice_type VARCHAR(20) DEFAULT 'standard',
    
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
    tax_type VARCHAR(20) NULL,
    
    -- 企业信息
    company_name VARCHAR(255) NULL,
    company_address TEXT NULL,
    company_tax_id VARCHAR(50) NULL,
    
    -- 发票内容
    description TEXT NOT NULL,
    line_items JSON NULL,
    
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
    
    -- 索引
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_invoice_number (invoice_number),
    INDEX idx_due_at (due_at),
    INDEX idx_created_at (created_at),
    INDEX idx_user_status_created (user_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='发票表';

-- 3. 创建支付表 - 资金流转记录
CREATE TABLE payments (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    payment_number VARCHAR(32) UNIQUE NOT NULL,
    invoice_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    
    -- 支付状态
    status VARCHAR(20) DEFAULT 'pending',
    payment_intent_id VARCHAR(255) NULL,
    
    -- 金额信息
    amount DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    exchange_rate DECIMAL(10,6) DEFAULT 1.0,
    
    -- 支付渠道信息
    payment_method VARCHAR(50) NOT NULL,
    payment_gateway VARCHAR(50) NOT NULL,
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
    
    -- 索引
    INDEX idx_invoice_id (invoice_id),
    INDEX idx_user_id (user_id),
    INDEX idx_status (status),
    INDEX idx_payment_method (payment_method),
    INDEX idx_payment_gateway (payment_gateway),
    INDEX idx_gateway_transaction_id (gateway_transaction_id),
    INDEX idx_payment_number (payment_number),
    INDEX idx_created_at (created_at),
    INDEX idx_user_status_created (user_id, status, created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='支付表';

-- 4. 创建订阅表 - 激活的服务实例
CREATE TABLE subscriptions (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    order_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    
    -- 订阅状态
    status VARCHAR(20) DEFAULT 'active',
    
    -- 服务时间
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    current_period_start TIMESTAMP NOT NULL,
    current_period_end TIMESTAMP NOT NULL,
    
    -- 计费配置
    billing_cycle VARCHAR(20) NOT NULL,
    billing_interval INT DEFAULT 1,
    price DECIMAL(10,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- 续费配置
    auto_renew BOOLEAN DEFAULT TRUE,
    next_billing_date TIMESTAMP NULL,
    
    -- 试用配置
    trial_end_date TIMESTAMP NULL,
    
    -- 取消配置
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancellation_reason TEXT NULL,
    cancelled_at TIMESTAMP NULL,
    
    -- 续费失败
    renewal_attempts INT DEFAULT 0,
    last_renewal_failed TIMESTAMP NULL,
    renewal_fail_reason TEXT NULL,
    
    -- 使用记录
    last_used_at TIMESTAMP NULL,
    
    -- 服务器组权限
    server_group_ids JSON NULL,
    
    -- 业务字段
    notes TEXT NULL,
    metadata JSON NULL,
    
    -- 时间戳
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- 索引
    INDEX idx_uuid (uuid),
    INDEX idx_order_id (order_id),
    INDEX idx_user_id (user_id),
    INDEX idx_plan_id (plan_id),
    INDEX idx_status (status),
    INDEX idx_end_date (end_date),
    INDEX idx_next_billing_date (next_billing_date),
    INDEX idx_created_at (created_at),
    INDEX idx_user_status_end (user_id, status, end_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='订阅表';

-- 5. 插入系统配置
INSERT INTO settings (key_name, value, description) VALUES
('invoice_pdf_template', 'default', '默认发票PDF模板'),
('payment_timeout_minutes', '30', '支付超时时间（分钟）'),
('invoice_due_days', '30', '发票默认付款期限（天）'),
('auto_reminder_days', '3,7,14', '发票自动提醒天数'),
('currency_default', 'USD', '默认货币'),
('invoice_number_prefix', 'INV', '发票号前缀'),
('order_number_prefix', 'ORD', '订单号前缀'),
('payment_number_prefix', 'PAY', '支付号前缀')
ON DUPLICATE KEY UPDATE value = VALUES(value);