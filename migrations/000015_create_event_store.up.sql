-- Create event store table for event sourcing and audit trail
CREATE TABLE IF NOT EXISTS event_store (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    event_id VARCHAR(255) NOT NULL UNIQUE COMMENT 'Unique identifier for the event',
    event_type VARCHAR(100) NOT NULL COMMENT 'Type of the event (e.g., user.created, payment.completed)',
    event_source VARCHAR(100) NOT NULL COMMENT 'Service or domain that generated the event',
    aggregate_id VARCHAR(100) COMMENT 'ID of the aggregate that the event relates to',
    aggregate_type VARCHAR(50) COMMENT 'Type of the aggregate (e.g., user, payment, subscription)',
    event_version VARCHAR(20) NOT NULL DEFAULT '1.0' COMMENT 'Version of the event schema',
    event_data TEXT NOT NULL COMMENT 'JSON serialized event data',
    metadata TEXT COMMENT 'JSON serialized event metadata',
    occurred_at TIMESTAMP NOT NULL COMMENT 'Timestamp when the event originally occurred',
    stored_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'Timestamp when the event was stored in the database',
    
    -- Create indexes
    INDEX idx_event_store_event_type (event_type),
    INDEX idx_event_store_event_source (event_source),
    INDEX idx_event_store_aggregate (aggregate_id, aggregate_type),
    INDEX idx_event_store_occurred_at (occurred_at),
    INDEX idx_event_store_stored_at (stored_at),
    
    -- Create composite indexes for common queries
    INDEX idx_event_store_type_occurred (event_type, occurred_at),
    INDEX idx_event_store_source_occurred (event_source, occurred_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Event store for domain events, audit trail, and event sourcing';