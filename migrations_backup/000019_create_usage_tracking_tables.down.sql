-- Drop usage tracking and alert tables
-- This migration removes all usage tracking capabilities

-- Drop tables in reverse order of creation to handle any dependencies
DROP TABLE IF EXISTS usage_alerts;
DROP TABLE IF EXISTS alert_configurations;
DROP TABLE IF EXISTS usage_records;