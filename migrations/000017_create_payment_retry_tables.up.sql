-- Create payment retry tables for Smart Payment Retry Strategy

-- Payment retries table to track retry sequences for failed payments
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Retry Information
    attempt_number INT NOT NULL DEFAULT 0 COMMENT 'Current attempt number (0-based)',
    max_attempts INT NOT NULL DEFAULT 3 COMMENT 'Maximum retry attempts',
    next_retry_at TIMESTAMP NOT NULL COMMENT 'Next retry time',
    last_attempt_at TIMESTAMP NOT NULL COMMENT 'Last attempt time',
    retry_strategy VARCHAR(50) NOT NULL COMMENT 'Strategy type: exponential, linear, custom',
    
    -- Retry Configuration
    initial_delay INT NOT NULL DEFAULT 3600 COMMENT 'Initial delay in seconds (1 hour)',
    max_delay INT NOT NULL DEFAULT 86400 COMMENT 'Maximum delay in seconds (24 hours)',
    backoff_factor DECIMAL(4,2) NOT NULL DEFAULT 2.0 COMMENT 'Backoff multiplier',
    
    -- Status and State
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending, in_progress, completed, failed, cancelled',
    failure_type VARCHAR(30) NULL COMMENT 'temporary, permanent, network, gateway, business',
    last_failure_code VARCHAR(50) NULL COMMENT 'Last error/failure code',
    last_error_message VARCHAR(500) NULL COMMENT 'Last error message',
    
    -- Gateway-specific Configuration
    gateway_config TEXT NULL COMMENT 'JSON config for gateway-specific retry settings',
    
    -- Tracking Information
    total_delay_time INT DEFAULT 0 COMMENT 'Total time spent in retries (seconds)',
    completed_at TIMESTAMP NULL COMMENT 'When retry sequence completed',
    cancelled_at TIMESTAMP NULL COMMENT 'When retry sequence was cancelled',
    successful_at TIMESTAMP NULL COMMENT 'When payment finally succeeded',
    
    -- Metadata
    metadata TEXT NULL COMMENT 'Additional retry metadata (JSON)',
    notes VARCHAR(500) NULL COMMENT 'Admin notes',
    
    -- Standard timestamp fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_payment_retries_payment_record_id (payment_record_id),
    INDEX idx_payment_retries_status (status),
    INDEX idx_payment_retries_next_retry_at (next_retry_at),
    INDEX idx_payment_retries_failure_type (failure_type),
    INDEX idx_payment_retries_completed_at (completed_at),
    INDEX idx_payment_retries_cancelled_at (cancelled_at),
    INDEX idx_payment_retries_successful_at (successful_at),
    INDEX idx_payment_retries_created_at (created_at),
    INDEX idx_payment_retries_deleted_at (deleted_at),
    
    -- Foreign key constraint
    FOREIGN KEY (payment_record_id) REFERENCES payment_records(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment retry tracking table';

-- Payment retry history table to track individual retry attempts
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Attempt Information
    attempt_number INT NOT NULL COMMENT 'Which attempt this was',
    attempted_at TIMESTAMP NOT NULL COMMENT 'When this attempt was made',
    duration INT DEFAULT 0 COMMENT 'Duration of attempt in milliseconds',
    
    -- Result Information
    status VARCHAR(20) NOT NULL COMMENT 'success, failed, timeout, error',
    response_code VARCHAR(50) NULL COMMENT 'Gateway response code',
    response_message VARCHAR(500) NULL COMMENT 'Gateway response message',
    error_type VARCHAR(30) NULL COMMENT 'Type of error encountered',
    failure_reason VARCHAR(500) NULL COMMENT 'Detailed failure reason',
    
    -- Technical Details
    request_data TEXT NULL COMMENT 'Request sent to gateway (sanitized)',
    response_data TEXT NULL COMMENT 'Response from gateway (sanitized)',
    
    -- Next Retry Information
    next_retry_at TIMESTAMP NULL COMMENT 'When next retry is scheduled',
    delay_from_previous INT DEFAULT 0 COMMENT 'Delay from previous attempt (seconds)',
    
    -- Metadata
    metadata TEXT NULL COMMENT 'Additional attempt metadata (JSON)',
    
    -- Standard timestamp fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_payment_retry_histories_payment_retry_id (payment_retry_id),
    INDEX idx_payment_retry_histories_payment_record_id (payment_record_id),
    INDEX idx_payment_retry_histories_attempt_number (attempt_number),
    INDEX idx_payment_retry_histories_attempted_at (attempted_at),
    INDEX idx_payment_retry_histories_status (status),
    INDEX idx_payment_retry_histories_error_type (error_type),
    INDEX idx_payment_retry_histories_next_retry_at (next_retry_at),
    INDEX idx_payment_retry_histories_created_at (created_at),
    INDEX idx_payment_retry_histories_deleted_at (deleted_at),
    
    -- Foreign key constraints
    FOREIGN KEY (payment_retry_id) REFERENCES payment_retries(id) ON DELETE CASCADE,
    FOREIGN KEY (payment_record_id) REFERENCES payment_records(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment retry attempt history table';

-- Add fields to payment_records table for retry tracking
ALTER TABLE payment_records 
ADD COLUMN retry_count INT NOT NULL DEFAULT 0 COMMENT 'Number of times payment has been retried',
ADD COLUMN last_retry_at TIMESTAMP NULL COMMENT 'Last time payment was retried',
ADD COLUMN is_retry_enabled BOOLEAN NOT NULL DEFAULT TRUE COMMENT 'Whether retry is enabled for this payment',
ADD COLUMN retry_failure_reason VARCHAR(500) NULL COMMENT 'Last retry failure reason',
ADD INDEX idx_payment_records_retry_count (retry_count),
ADD INDEX idx_payment_records_last_retry_at (last_retry_at),
ADD INDEX idx_payment_records_is_retry_enabled (is_retry_enabled);

-- Create a view for active payment retries (pending and scheduled)
CREATE VIEW v_active_payment_retries AS
SELECT 
    pr.id,
    pr.payment_record_id,
    pr.attempt_number,
    pr.max_attempts,
    pr.next_retry_at,
    pr.last_attempt_at,
    pr.retry_strategy,
    pr.status,
    pr.failure_type,
    pr.last_failure_code,
    pr.last_error_message,
    pr.created_at,
    p.payment_no,
    p.gateway,
    p.payment_method,
    p.amount,
    p.currency,
    p.user_id
FROM payment_retries pr
JOIN payment_records p ON pr.payment_record_id = p.id
WHERE pr.status IN ('pending', 'in_progress')
  AND pr.deleted_at IS NULL
  AND p.deleted_at IS NULL;

-- Create a view for retry statistics by gateway
CREATE VIEW v_payment_retry_stats AS
SELECT 
    p.gateway,
    p.payment_method,
    COUNT(DISTINCT pr.id) as total_retries,
    COUNT(DISTINCT CASE WHEN pr.status = 'completed' THEN pr.id END) as successful_retries,
    COUNT(DISTINCT CASE WHEN pr.status = 'failed' THEN pr.id END) as failed_retries,
    COUNT(DISTINCT CASE WHEN pr.status = 'cancelled' THEN pr.id END) as cancelled_retries,
    AVG(pr.attempt_number) as avg_attempts,
    AVG(pr.total_delay_time) as avg_delay_time,
    DATE(pr.created_at) as retry_date
FROM payment_retries pr
JOIN payment_records p ON pr.payment_record_id = p.id
WHERE pr.deleted_at IS NULL
  AND p.deleted_at IS NULL
GROUP BY p.gateway, p.payment_method, DATE(pr.created_at);