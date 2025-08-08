-- 为用户第三方账号绑定表添加软删除支持
ALTER TABLE user_account_bindings 
ADD COLUMN deleted_at TIMESTAMP NULL COMMENT '软删除时间戳';

-- 为 deleted_at 字段添加索引以提升查询性能
CREATE INDEX idx_deleted_at ON user_account_bindings (deleted_at) COMMENT '软删除索引';