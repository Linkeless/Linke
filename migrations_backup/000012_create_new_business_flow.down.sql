-- 回滚新商务流程数据库结构

-- 1. 删除主要业务表（按依赖关系顺序）
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS orders;

-- 2. 删除配置数据
DELETE FROM settings WHERE key_name IN (
    'invoice_pdf_template',
    'payment_timeout_minutes',
    'invoice_due_days',
    'auto_reminder_days',
    'currency_default',
    'invoice_number_prefix',
    'order_number_prefix',
    'payment_number_prefix'
);