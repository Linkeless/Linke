-- Complete Database Schema Rollback
-- This migration removes all tables and data

-- Drop indexes first
DROP INDEX IF EXISTS idx_login_attempts_ip_analysis ON login_attempts;
DROP INDEX IF EXISTS idx_login_attempts_analysis ON login_attempts;
DROP INDEX IF EXISTS idx_jwt_blacklist_cleanup ON jwt_blacklist;

-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS account_lockouts;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS jwt_blacklist;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS user_subscription_server_groups;
DROP TABLE IF EXISTS user_traffic_logs;
DROP TABLE IF EXISTS node_data;
DROP TABLE IF EXISTS payment_configs;
DROP TABLE IF EXISTS shadowsocks_servers;
DROP TABLE IF EXISTS server_groups;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS referral_campaigns;
DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS subscription_orders;
DROP TABLE IF EXISTS payment_records;
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS subscription_plans;
DROP TABLE IF EXISTS invite_code_usages;
DROP TABLE IF EXISTS invite_codes;
DROP TABLE IF EXISTS users;