-- Remove default server group IDs field from subscription_plans table
ALTER TABLE subscription_plans 
DROP COLUMN default_server_group_ids;