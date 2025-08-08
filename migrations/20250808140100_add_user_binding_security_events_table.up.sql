-- 创建用户绑定安全事件表
CREATE TABLE user_binding_security_events (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL COMMENT '事件类型：suspicious_binding, duplicate_attempt, rapid_binding等',
    severity VARCHAR(20) NOT NULL COMMENT '严重级别：low, medium, high, critical',
    user_id INT UNSIGNED NULL COMMENT '相关的用户ID（可能为空）',
    provider VARCHAR(20) NOT NULL COMMENT '第三方提供商: google, github, telegram',
    provider_user_id VARCHAR(100) NOT NULL COMMENT '第三方平台的用户ID',
    ip_address VARCHAR(45) COMMENT '事件发生的IP地址（支持IPv6）',
    user_agent VARCHAR(500) COMMENT '用户代理字符串',
    description TEXT NOT NULL COMMENT '事件详细描述',
    metadata JSON COMMENT '事件相关的额外元数据',
    resolved BOOLEAN DEFAULT FALSE COMMENT '是否已解决',
    resolved_by INT UNSIGNED NULL COMMENT '解决事件的管理员ID',
    resolved_at TIMESTAMP NULL COMMENT '解决时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '事件发生时间',
    
    -- 索引优化
    INDEX idx_event_type (event_type) COMMENT '事件类型索引',
    INDEX idx_severity (severity) COMMENT '严重级别索引',
    INDEX idx_user_id (user_id) COMMENT '用户查询索引',
    INDEX idx_provider (provider) COMMENT '提供商索引',
    INDEX idx_provider_user_id (provider_user_id) COMMENT '第三方用户ID索引',
    INDEX idx_resolved (resolved) COMMENT '解决状态索引',
    INDEX idx_resolved_by (resolved_by) COMMENT '解决人索引',
    INDEX idx_resolved_at (resolved_at) COMMENT '解决时间索引',
    INDEX idx_created_at (created_at) COMMENT '创建时间索引',
    INDEX idx_ip_address (ip_address) COMMENT 'IP地址索引',
    INDEX idx_severity_created (severity, created_at) COMMENT '严重级别和时间复合索引',
    INDEX idx_provider_severity (provider, severity) COMMENT '提供商和严重级别复合索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户绑定安全事件表';