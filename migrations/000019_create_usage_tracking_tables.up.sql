-- Create usage tracking and alert tables for real-time monitoring
-- This migration adds comprehensive usage tracking capabilities including:
-- - Detailed usage records for time-series data
-- - Alert configurations for threshold-based monitoring
-- - Usage alerts for notifications and tracking
-- - Optimized indexes for high-performance queries

-- ==============================================================================
-- USAGE RECORDS TABLE - TIME-SERIES USAGE DATA
-- ==============================================================================

CREATE TABLE usage_records (
    -- Primary Key
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Foreign Key
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Usage Details
    usage_type VARCHAR(50) NOT NULL COMMENT 'Type of usage (traffic, api_calls, storage, etc.)',
    amount BIGINT NOT NULL COMMENT 'Usage amount in bytes or count',
    unit VARCHAR(20) NOT NULL DEFAULT 'bytes' COMMENT 'Unit of measurement (bytes, count, etc.)',
    timestamp TIMESTAMP NOT NULL COMMENT 'When the usage occurred',
    
    -- Source Information
    source_type VARCHAR(50) NOT NULL COMMENT 'Source of usage (server, api, admin, etc.)',
    source_id VARCHAR(100) NULL COMMENT 'ID of the source (server_id, api_endpoint, etc.)',
    
    -- Additional Context
    metadata TEXT NULL COMMENT 'Additional metadata as JSON',
    user_agent TEXT NULL COMMENT 'User agent if applicable',
    ip_address VARCHAR(45) NULL COMMENT 'IP address if applicable',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Performance indexes for time-series queries
    INDEX idx_usage_records_subscription (user_subscription_id, usage_type, timestamp),
    INDEX idx_usage_records_timestamp (timestamp),
    INDEX idx_usage_records_usage_type (usage_type, timestamp),
    INDEX idx_usage_records_source (source_type, source_id),
    INDEX idx_usage_records_ip (ip_address),
    INDEX idx_usage_records_deleted (deleted_at),
    
    -- Composite indexes for common aggregation queries
    INDEX idx_usage_records_subscription_type_time (user_subscription_id, usage_type, timestamp DESC),
    INDEX idx_usage_records_hourly_agg (user_subscription_id, usage_type, DATE_FORMAT(timestamp, '%Y-%m-%d %H')),
    INDEX idx_usage_records_daily_agg (user_subscription_id, usage_type, DATE(timestamp)),
    
    -- Covering index for reporting queries
    INDEX idx_usage_records_covering (user_subscription_id, usage_type, timestamp, amount, unit)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Time-series usage tracking data for real-time monitoring';

-- ==============================================================================
-- ALERT CONFIGURATIONS TABLE - USER-DEFINED ALERT THRESHOLDS
-- ==============================================================================

CREATE TABLE alert_configurations (
    -- Primary Key
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Foreign Key
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Settings
    usage_type VARCHAR(50) NOT NULL COMMENT 'Type of usage to monitor',
    threshold_type VARCHAR(20) NOT NULL DEFAULT 'percentage' COMMENT 'Type of threshold (percentage, absolute)',
    threshold DECIMAL(10,4) NOT NULL COMMENT 'Alert threshold value',
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether alert is enabled',
    
    -- Notification Settings
    notification_channels TEXT NULL COMMENT 'JSON array of notification channels',
    cooldown_minutes INT NOT NULL DEFAULT 60 COMMENT 'Cooldown period between alerts in minutes',
    
    -- Metadata
    name VARCHAR(100) NOT NULL COMMENT 'Human readable name for the alert',
    description TEXT NULL COMMENT 'Description of what this alert monitors',
    priority VARCHAR(20) NOT NULL DEFAULT 'medium' COMMENT 'Alert priority level',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_alert_configs_subscription (user_subscription_id, is_enabled),
    INDEX idx_alert_configs_usage_type (usage_type, is_enabled),
    INDEX idx_alert_configs_enabled (is_enabled, deleted_at),
    INDEX idx_alert_configs_priority (priority),
    INDEX idx_alert_configs_deleted (deleted_at),
    
    -- Composite index for active alert lookup
    INDEX idx_alert_configs_active (user_subscription_id, usage_type, is_enabled, deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='User-configured alert thresholds for usage monitoring';

-- ==============================================================================
-- USAGE ALERTS TABLE - FIRED ALERTS AND NOTIFICATION TRACKING
-- ==============================================================================

CREATE TABLE usage_alerts (
    -- Primary Key
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    
    -- Foreign Keys
    user_subscription_id BIGINT UNSIGNED NOT NULL,
    alert_configuration_id BIGINT UNSIGNED NOT NULL,
    
    -- Alert Details
    usage_type VARCHAR(50) NOT NULL,
    current_usage BIGINT NOT NULL COMMENT 'Current usage amount when alert fired',
    usage_limit BIGINT NOT NULL COMMENT 'Usage limit at time of alert',
    threshold_value DECIMAL(10,4) NOT NULL COMMENT 'Threshold that was exceeded',
    usage_percent DECIMAL(5,2) NOT NULL COMMENT 'Percentage of limit used',
    
    -- Alert State
    status VARCHAR(20) NOT NULL DEFAULT 'fired' COMMENT 'Alert status',
    severity VARCHAR(20) NOT NULL COMMENT 'Alert severity level',
    fired_at TIMESTAMP NOT NULL COMMENT 'When the alert was first fired',
    resolved_at TIMESTAMP NULL COMMENT 'When the alert was resolved',
    
    -- Notification Tracking
    notifications_sent INT NOT NULL DEFAULT 0 COMMENT 'Number of notifications sent',
    last_notification_sent TIMESTAMP NULL COMMENT 'When last notification was sent',
    notification_channels TEXT NULL COMMENT 'Channels used for notifications',
    notification_results TEXT NULL COMMENT 'Results of notification attempts as JSON',
    
    -- Additional Context
    message TEXT NULL COMMENT 'Alert message',
    metadata TEXT NULL COMMENT 'Additional metadata as JSON',
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    PRIMARY KEY (id),
    
    -- Essential indexes
    INDEX idx_usage_alerts_subscription (user_subscription_id, status),
    INDEX idx_usage_alerts_config (alert_configuration_id),
    INDEX idx_usage_alerts_usage_type (usage_type, status),
    INDEX idx_usage_alerts_status (status, resolved_at),
    INDEX idx_usage_alerts_fired_at (fired_at),
    INDEX idx_usage_alerts_resolved_at (resolved_at),
    INDEX idx_usage_alerts_severity (severity, status),
    INDEX idx_usage_alerts_deleted (deleted_at),
    
    -- Composite indexes for dashboard queries
    INDEX idx_usage_alerts_active (user_subscription_id, status, fired_at DESC),
    INDEX idx_usage_alerts_recent (status, fired_at DESC, resolved_at),
    INDEX idx_usage_alerts_notifications (last_notification_sent, status),
    
    -- Covering index for alert list queries
    INDEX idx_usage_alerts_covering (user_subscription_id, status, fired_at, severity, usage_type)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
COMMENT='Fired usage alerts and notification tracking';

-- ==============================================================================
-- DEFAULT ALERT CONFIGURATIONS
-- ==============================================================================

-- Insert default alert configurations for common thresholds
-- These will be used as templates when users create new subscriptions

-- Traffic usage alerts (50%, 80%, 90%, 100%)
INSERT INTO alert_configurations (
    user_subscription_id, usage_type, threshold_type, threshold, 
    name, description, priority, notification_channels, cooldown_minutes
) VALUES 
-- Note: user_subscription_id = 0 will be used as templates
-- These will be copied for each new subscription
(0, 'traffic', 'percentage', 50.0, 'Traffic 50% Warning', 'Alert when traffic usage reaches 50% of limit', 'low', 
 '[{"type":"in_app","target":"user","enabled":true}]', 1440), -- 24 hours cooldown

(0, 'traffic', 'percentage', 80.0, 'Traffic 80% Alert', 'Alert when traffic usage reaches 80% of limit', 'medium', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 720), -- 12 hours cooldown

(0, 'traffic', 'percentage', 90.0, 'Traffic 90% Critical', 'Critical alert when traffic usage reaches 90% of limit', 'high', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 360), -- 6 hours cooldown

(0, 'traffic', 'percentage', 100.0, 'Traffic Limit Exceeded', 'Critical alert when traffic limit is exceeded', 'critical', 
 '[{"type":"email","target":"user","enabled":true},{"type":"in_app","target":"user","enabled":true}]', 60); -- 1 hour cooldown

-- ==============================================================================
-- INDEXES FOR PERFORMANCE OPTIMIZATION
-- ==============================================================================

-- Additional composite indexes for complex queries

-- Usage records aggregation by hour
CREATE INDEX idx_usage_records_hourly_stats ON usage_records (
    user_subscription_id, 
    usage_type, 
    YEAR(timestamp), 
    MONTH(timestamp), 
    DAY(timestamp), 
    HOUR(timestamp)
);

-- Usage records aggregation by day
CREATE INDEX idx_usage_records_daily_stats ON usage_records (
    user_subscription_id, 
    usage_type, 
    DATE(timestamp)
);

-- Alert configurations for quick lookup during usage updates
CREATE INDEX idx_alert_configs_threshold_lookup ON alert_configurations (
    user_subscription_id, 
    usage_type, 
    is_enabled, 
    threshold_type, 
    threshold
);

-- Usage alerts for dashboard and reporting
CREATE INDEX idx_usage_alerts_dashboard ON usage_alerts (
    user_subscription_id, 
    status, 
    severity, 
    fired_at DESC
);

-- ==============================================================================
-- PARTITIONING PREPARATION (Optional - for high-volume deployments)
-- ==============================================================================

-- For high-volume deployments, consider partitioning usage_records by date
-- This is commented out as it requires careful planning and may not be needed initially

/*
-- Example: Partition usage_records by month
ALTER TABLE usage_records 
PARTITION BY RANGE (YEAR(timestamp) * 100 + MONTH(timestamp)) (
    PARTITION p202401 VALUES LESS THAN (202402),
    PARTITION p202402 VALUES LESS THAN (202403),
    PARTITION p202403 VALUES LESS THAN (202404),
    -- Add more partitions as needed
    PARTITION p_future VALUES LESS THAN MAXVALUE
);
*/

-- ==============================================================================
-- PERFORMANCE HINTS AND NOTES
-- ==============================================================================

/*
Performance optimization notes:

1. Usage Records Table:
   - Designed for high-frequency inserts with minimal locking
   - Composite indexes support efficient range queries for time-series data
   - Consider archiving old records (older than 1 year) to maintain performance

2. Alert Configurations:
   - Small table with efficient lookups during usage updates
   - Template records (user_subscription_id = 0) for easy default setup

3. Usage Alerts:
   - Optimized for dashboard queries and notification processing
   - Status-based indexes for quick filtering of active alerts

4. Query Patterns:
   - Real-time usage: Use covering indexes for subscription + type + time ranges
   - Aggregations: Use date-based composite indexes
   - Alert processing: Use threshold lookup indexes

5. Maintenance:
   - Schedule regular cleanup of old usage_records (>1 year)
   - Monitor index usage and adjust as query patterns evolve
   - Consider read replicas for reporting workloads
*/