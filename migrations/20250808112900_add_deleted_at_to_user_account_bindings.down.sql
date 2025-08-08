-- 移除用户第三方账号绑定表的软删除支持
DROP INDEX idx_deleted_at ON user_account_bindings;

ALTER TABLE user_account_bindings 
DROP COLUMN deleted_at;