-- ==============================================================================
-- COMPREHENSIVE DATABASE SCHEMA ROLLBACK - CONSOLIDATED MIGRATION
-- ==============================================================================
-- This rollback migration safely removes all database structures created by
-- the corresponding up migration. Order is important to handle dependencies.
-- ==============================================================================

-- Drop all tables in reverse dependency order

-- ==============================================================================
-- SYSTEM AND CONFIGURATION TABLES
-- ==============================================================================

DROP TABLE IF EXISTS settings;

-- ==============================================================================
-- EVENT STORE
-- ==============================================================================

DROP TABLE IF EXISTS event_store;

-- ==============================================================================
-- TRAFFIC MANAGEMENT SYSTEM
-- ==============================================================================

DROP TABLE IF EXISTS user_traffic_logs;
DROP TABLE IF EXISTS node_data;

-- ==============================================================================
-- SUPPORT SYSTEM
-- ==============================================================================

DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;

-- ==============================================================================
-- SERVER MANAGEMENT
-- ==============================================================================

DROP TABLE IF EXISTS user_subscription_server_groups;
DROP TABLE IF EXISTS shadowsocks_servers;
DROP TABLE IF EXISTS server_groups;

-- ==============================================================================
-- REFERRAL SYSTEM
-- ==============================================================================

DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS referral_campaigns;
DROP TABLE IF EXISTS invite_code_usages;
DROP TABLE IF EXISTS invite_codes;

-- ==============================================================================
-- COUPON SYSTEM
-- ==============================================================================

DROP TABLE IF EXISTS coupon_usages;
DROP TABLE IF EXISTS coupons;

-- ==============================================================================
-- USAGE TRACKING AND ALERTS
-- ==============================================================================

DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS alert_configurations;
DROP TABLE IF EXISTS usage_records;

-- ==============================================================================
-- INVOICE MANAGEMENT
-- ==============================================================================

DROP TABLE IF EXISTS invoices;

-- ==============================================================================
-- PAYMENT PROCESSING
-- ==============================================================================

-- Drop payment retry tables first (dependent on payment_records)
DROP TABLE IF EXISTS payment_retry_histories;
DROP TABLE IF EXISTS payment_retries;

-- Drop payment configuration and methods
DROP TABLE IF EXISTS payment_configs;
DROP TABLE IF EXISTS payment_methods;

-- Drop main payment records table
DROP TABLE IF EXISTS payment_records;

-- ==============================================================================
-- ORDER AND BILLING SYSTEM
-- ==============================================================================

DROP TABLE IF EXISTS subscription_orders;

-- ==============================================================================
-- SUBSCRIPTION MANAGEMENT
-- ==============================================================================

-- Drop user subscriptions first (dependent on subscription_plans)
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS subscription_plans;

-- ==============================================================================
-- SECURITY AND AUTHENTICATION
-- ==============================================================================

DROP TABLE IF EXISTS account_lockouts;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS jwt_blacklist;

-- ==============================================================================
-- CORE USER MANAGEMENT
-- ==============================================================================

-- Drop users table last as it's referenced by many other tables
DROP TABLE IF EXISTS users;

-- ==============================================================================
-- ROLLBACK CONFIRMATION
-- ==============================================================================

-- The following comment confirms successful rollback
-- All tables, indexes, and constraints have been removed
-- Database schema has been rolled back to pre-migration state