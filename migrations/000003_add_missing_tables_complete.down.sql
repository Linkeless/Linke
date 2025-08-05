-- ==============================================================================
-- ADD MISSING TABLES - COMPLETE MIGRATION DOWN
-- ==============================================================================
-- This down migration removes all 6 missing tables that were added
-- ==============================================================================

-- Drop tables in reverse order to handle any logical dependencies
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS alert_configurations;
DROP TABLE IF EXISTS referral_campaigns;
DROP TABLE IF EXISTS invite_codes;
DROP TABLE IF EXISTS usage_records;
DROP TABLE IF EXISTS referrals;