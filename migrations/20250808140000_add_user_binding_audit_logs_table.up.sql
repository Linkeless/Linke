-- 创建用户绑定操作审计日志表
CREATE TABLE user_binding_audit_logs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL COMMENT '操作的用户ID',
    admin_id INT UNSIGNED NULL COMMENT '执行操作的管理员ID（如果是管理员操作）',
    binding_id INT UNSIGNED NULL COMMENT '相关的绑定ID（可能为空，如批量操作）',
    operation VARCHAR(50) NOT NULL COMMENT '操作类型：create, update, delete, batch_create等',
    provider VARCHAR(20) NOT NULL COMMENT '第三方提供商: google, github, telegram',
    details TEXT COMMENT '操作详细信息的JSON格式',
    ip_address VARCHAR(45) COMMENT '操作的IP地址（支持IPv6）',
    user_agent VARCHAR(500) COMMENT '用户代理字符串',
    status VARCHAR(20) NOT NULL COMMENT '操作状态：success, failure, warning',
    error_message TEXT NULL COMMENT '错误信息（如果操作失败）',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
    
    -- 索引优化
    INDEX idx_user_id (user_id) COMMENT '用户查询索引',
    INDEX idx_admin_id (admin_id) COMMENT '管理员查询索引',
    INDEX idx_binding_id (binding_id) COMMENT '绑定查询索引',
    INDEX idx_operation (operation) COMMENT '操作类型索引',
    INDEX idx_provider (provider) COMMENT '提供商索引',
    INDEX idx_status (status) COMMENT '状态索引',
    INDEX idx_created_at (created_at) COMMENT '时间查询索引',
    INDEX idx_user_operation_time (user_id, operation, created_at) COMMENT '用户操作时间复合索引'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户绑定操作审计日志表';