-- ==============================================================================
-- ROLLBACK: REMOVE REMAINING MISSING TABLES
-- ==============================================================================
-- This migration rolls back the creation of the 5 tables added in the up migration
-- ==============================================================================

-- Drop tables in reverse order to handle any dependencies
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS alert_configurations;
DROP TABLE IF EXISTS referral_campaigns;
DROP TABLE IF EXISTS invite_codes;
DROP TABLE IF EXISTS usage_records;