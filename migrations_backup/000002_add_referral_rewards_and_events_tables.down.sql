-- ==============================================================================
-- ADD MISSING REFERRAL TABLES - DOWN MIGRATION
-- ==============================================================================
-- This migration removes the referral_rewards and referral_events tables
-- ==============================================================================

-- Drop referral events table
DROP TABLE IF EXISTS referral_events;

-- Drop referral rewards table  
DROP TABLE IF EXISTS referral_rewards;