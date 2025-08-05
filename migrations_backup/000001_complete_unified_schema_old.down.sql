-- ==============================================================================
-- COMPLETE UNIFIED SCHEMA MIGRATION - DOWN
-- ==============================================================================
-- This down migration removes all tables created by the unified schema migration
-- Tables are dropped in reverse dependency order
-- ==============================================================================

-- Drop tables in reverse dependency order to handle logical dependencies

-- Event sourcing
DROP TABLE IF EXISTS event_store;

-- Traffic management
DROP TABLE IF EXISTS user_traffic_logs;
DROP TABLE IF EXISTS node_data;

-- Support system
DROP TABLE IF EXISTS ticket_messages;
DROP TABLE IF EXISTS tickets;

-- Server management
DROP TABLE IF EXISTS user_subscription_server_groups;
DROP TABLE IF EXISTS shadowsocks_servers;
DROP TABLE IF EXISTS server_groups;

-- Complete referral system
DROP TABLE IF EXISTS referral_events;
DROP TABLE IF EXISTS referral_rewards;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS referral_campaigns;
DROP TABLE IF EXISTS invite_code_usages;
DROP TABLE IF EXISTS invite_codes;

-- Coupon system
DROP TABLE IF EXISTS coupon_usages;
DROP TABLE IF EXISTS coupons;

-- Usage tracking and monitoring
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS alert_configurations;
DROP TABLE IF EXISTS usage_records;

-- Invoice management
DROP TABLE IF EXISTS invoices;

-- Payment processing system
DROP TABLE IF EXISTS payment_configs;
DROP TABLE IF EXISTS payment_methods;
DROP TABLE IF EXISTS payment_retry_histories;
DROP TABLE IF EXISTS payment_retries;
DROP TABLE IF EXISTS payment_records;

-- Order and billing management
DROP TABLE IF EXISTS subscription_orders;

-- Subscription management
DROP TABLE IF EXISTS user_subscriptions;
DROP TABLE IF EXISTS subscription_plans;

-- Authentication and security
DROP TABLE IF EXISTS account_lockouts;
DROP TABLE IF EXISTS login_attempts;
DROP TABLE IF EXISTS jwt_blacklist;

-- User management
DROP TABLE IF EXISTS users;

-- Core system
DROP TABLE IF EXISTS settings;