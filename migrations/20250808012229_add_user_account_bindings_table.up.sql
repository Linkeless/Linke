-- 创建用户第三方账号绑定表
CREATE TABLE user_account_bindings (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    user_id INT UNSIGNED NOT NULL COMMENT '关联的用户ID，无外键约束但应用层保证数据一致性',
    provider VARCHAR(20) NOT NULL COMMENT '第三方提供商: google, github, telegram',
    provider_user_id VARCHAR(100) NOT NULL COMMENT '第三方平台的用户ID',
    provider_email VARCHAR(255) COMMENT '第三方平台的邮箱',
    provider_username VARCHAR(100) COMMENT '第三方平台的用户名',
    provider_name VARCHAR(255) COMMENT '第三方平台的显示名称',
    provider_avatar VARCHAR(500) COMMENT '第三方平台的头像链接',
    provider_data JSON COMMENT '第三方平台的额外数据',
    is_primary BOOLEAN DEFAULT FALSE COMMENT '是否为主要绑定账号',
    bound_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '绑定时间',
    last_used_at TIMESTAMP NULL COMMENT '最后使用时间',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    
    -- 索引优化
    INDEX idx_user_id (user_id) COMMENT '用户查询索引',
    INDEX idx_provider_user (provider, provider_user_id) COMMENT '第三方账号查询索引',
    INDEX idx_provider_email (provider, provider_email) COMMENT '第三方邮箱查询索引',
    INDEX idx_last_used_at (last_used_at) COMMENT '最后使用时间排序索引',
    INDEX idx_created_at (created_at) COMMENT '创建时间排序索引',
    
    -- 唯一约束
    UNIQUE KEY unique_provider_user (provider, provider_user_id) COMMENT '确保第三方账号唯一',
    UNIQUE KEY unique_user_provider (user_id, provider) COMMENT '确保用户每个平台只能绑定一个账号'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户第三方账号绑定表';

-- 迁移现有用户的第三方账号绑定数据
INSERT INTO user_account_bindings (user_id, provider, provider_user_id, provider_email, bound_at, created_at, updated_at)
SELECT 
    id as user_id,
    'google' as provider,
    google_id as provider_user_id,
    email as provider_email,
    created_at as bound_at,
    created_at,
    updated_at
FROM users 
WHERE google_id IS NOT NULL AND google_id != '';

INSERT INTO user_account_bindings (user_id, provider, provider_user_id, provider_email, bound_at, created_at, updated_at)
SELECT 
    id as user_id,
    'github' as provider,
    github_id as provider_user_id,
    email as provider_email,
    created_at as bound_at,
    created_at,
    updated_at
FROM users 
WHERE github_id IS NOT NULL AND github_id != '';

INSERT INTO user_account_bindings (user_id, provider, provider_user_id, bound_at, created_at, updated_at)
SELECT 
    id as user_id,
    'telegram' as provider,
    telegram_id as provider_user_id,
    created_at as bound_at,
    created_at,
    updated_at
FROM users 
WHERE telegram_id IS NOT NULL AND telegram_id != '';

-- 为迁移的数据设置主要绑定（使用临时表避免MySQL限制）
UPDATE user_account_bindings 
JOIN (
    SELECT user_id, MIN(provider) as first_provider
    FROM user_account_bindings 
    GROUP BY user_id
) AS first_bindings ON user_account_bindings.user_id = first_bindings.user_id 
    AND user_account_bindings.provider = first_bindings.first_provider
SET user_account_bindings.is_primary = TRUE;