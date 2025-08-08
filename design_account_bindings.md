# 用户第三方账号绑定设计方案

## 设计目标
- 用户可以绑定多个第三方账号（Google、GitHub、Telegram）
- 绑定后可通过任意已绑定的账号登录
- 遵循 MySQL 最佳实践，不使用外键约束
- 保持向后兼容性，不破坏现有用户表

## 数据库表设计 (MySQL 最佳实践)

### 新表：user_account_bindings

```sql
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
```

### MySQL 最佳实践应用

1. **无外键约束**: 
   - 避免锁定和性能问题
   - 应用层负责数据一致性检查
   - 提高并发性能和扩展性

2. **索引优化**:
   - 复合索引优先级考虑查询频率
   - 避免过多单列索引
   - 使用覆盖索引提高查询性能

3. **数据类型优化**:
   - 使用 `INT UNSIGNED` 节省存储空间
   - `VARCHAR` 长度根据实际需求设定
   - `JSON` 类型存储动态数据

4. **字符集和排序规则**:
   - 使用 `utf8mb4` 支持完整 Unicode
   - `utf8mb4_unicode_ci` 提供更好的排序支持

5. **存储引擎**:
   - InnoDB 支持事务和行级锁定
   - 更好的崩溃恢复能力

### 字段说明
- `id`: 主键，自增
- `user_id`: 关联的用户ID（无外键，应用层维护一致性）
- `provider`: 第三方提供商（google, github, telegram）
- `provider_user_id`: 第三方平台的用户ID
- `provider_email`: 第三方平台的邮箱
- `provider_username`: 第三方平台的用户名
- `provider_name`: 第三方平台的显示名称
- `provider_avatar`: 第三方平台的头像链接
- `provider_data`: 第三方平台的额外数据（JSON格式）
- `is_primary`: 是否为主要绑定账号
- `bound_at`: 绑定时间
- `last_used_at`: 最后使用时间

### 数据一致性保证

**应用层约束**:
1. 绑定前检查 `user_id` 是否存在于 `users` 表
2. 删除用户时，应用层负责清理相关绑定记录
3. 事务处理确保操作的原子性

**业务规则**:
- 同一个第三方账号只能绑定一个用户
- 一个用户在同一个平台只能绑定一个账号
- 删除用户时清理所有相关绑定

## API设计

### 绑定接口
- `POST /api/v1/user/bindings/{provider}` - 绑定第三方账号
- `GET /api/v1/user/bindings` - 获取当前用户的所有绑定
- `DELETE /api/v1/user/bindings/{provider}` - 解绑第三方账号
- `PUT /api/v1/user/bindings/{provider}/primary` - 设置主要绑定账号

### 登录支持
- 现有的 OAuth 登录接口支持通过绑定账号登录
- 如果第三方账号已绑定，则登录对应用户
- 如果未绑定，提示用户绑定或创建新账号

## 迁移策略

### 现有数据处理
1. 对于已有用户的 GoogleID、GitHubID、TelegramID
2. 创建对应的绑定记录到 user_account_bindings 表
3. 保留原有字段以确保兼容性

### 兼容性
- 保持现有 OAuth 登录流程不变
- 新增绑定功能为额外功能
- 现有用户数据自动迁移到新表结构

### 性能考虑
- 分页查询绑定记录
- 缓存活跃用户的绑定信息
- 定期清理过期的绑定记录