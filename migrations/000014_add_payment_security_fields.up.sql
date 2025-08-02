-- Add security enhancement fields to payment_records table
ALTER TABLE payment_records 
ADD COLUMN last_notify_time TIMESTAMP NULL COMMENT '最后一次通知时间，用于时间窗口验证',
ADD COLUMN notify_source VARCHAR(45) NULL COMMENT '通知来源IP，用于异常检测',
ADD INDEX idx_payment_records_last_notify_time (last_notify_time),
ADD INDEX idx_payment_records_notify_source (notify_source);