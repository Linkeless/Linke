-- 创建绑定验证配置表
CREATE TABLE user_binding_validation_configs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    max_bindings_per_user INT NOT NULL DEFAULT 10 COMMENT '每个用户最大绑定数量',
    max_bindings_per_provider INT NOT NULL DEFAULT 1 COMMENT '每个提供商的最大绑定数量',
    require_email_verification BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否要求邮箱验证',
    block_suspicious_providers BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否阻止可疑提供商',
    enable_rate_limit BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否启用频率限制',
    rate_limit_window_minutes INT NOT NULL DEFAULT 60 COMMENT '频率限制窗口（分钟）',
    rate_limit_max_attempts INT NOT NULL DEFAULT 5 COMMENT '频率限制最大尝试次数',
    security_score_threshold DECIMAL(5,2) NOT NULL DEFAULT 70.00 COMMENT '安全分数阈值（0-100）',
    config_name VARCHAR(50) NOT NULL DEFAULT 'default' COMMENT '配置名称',
    is_active BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否激活',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    created_by INT UNSIGNED NULL COMMENT '创建者ID',
    updated_by INT UNSIGNED NULL COMMENT '更新者ID',
    
    -- 索引优化
    INDEX idx_config_name (config_name) COMMENT '配置名称索引',
    INDEX idx_is_active (is_active) COMMENT '激活状态索引',
    INDEX idx_created_by (created_by) COMMENT '创建者索引',
    INDEX idx_updated_by (updated_by) COMMENT '更新者索引',
    
    -- 唯一约束
    UNIQUE KEY unique_active_config (config_name, is_active) COMMENT '确保只有一个激活的配置'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='绑定验证配置表';

-- 插入默认配置
INSERT INTO user_binding_validation_configs (
    config_name, max_bindings_per_user, max_bindings_per_provider, 
    require_email_verification, block_suspicious_providers, 
    enable_rate_limit, rate_limit_window_minutes, rate_limit_max_attempts,
    security_score_threshold, is_active
) VALUES (
    'default', 10, 1, TRUE, TRUE, TRUE, 60, 5, 70.00, TRUE
);