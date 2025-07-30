-- Add default server group IDs field to subscription_plans table
ALTER TABLE subscription_plans 
ADD COLUMN default_server_group_ids TEXT NULL 
COMMENT 'Default server groups for subscriptions (JSON)' 
AFTER traffic_reset_cycle;