-- Create payment retry tables for Smart Payment Retry Strategy (Simplified)

-- Payment retries table to track retry sequences for failed payments
CREATE TABLE payment_retries (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Retry Information
    attempt_number INT NOT NULL DEFAULT 0 COMMENT 'Current attempt number (0-based)',
    max_attempts INT NOT NULL DEFAULT 3 COMMENT 'Maximum retry attempts',
    next_retry_at TIMESTAMP NULL COMMENT 'Next retry time',
    last_attempt_at TIMESTAMP NULL COMMENT 'Last attempt time',
    retry_strategy VARCHAR(50) NOT NULL DEFAULT 'exponential' COMMENT 'Strategy type: exponential, linear, custom',
    
    -- Status and State
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT 'pending, in_progress, completed, failed, cancelled',
    failure_type VARCHAR(30) NULL COMMENT 'temporary, permanent, network, gateway, business',
    last_failure_code VARCHAR(50) NULL COMMENT 'Last error/failure code',
    last_error_message VARCHAR(500) NULL COMMENT 'Last error message',
    
    -- Standard timestamp fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_payment_retries_payment_record_id (payment_record_id),
    INDEX idx_payment_retries_status (status),
    INDEX idx_payment_retries_next_retry_at (next_retry_at),
    INDEX idx_payment_retries_created_at (created_at),
    INDEX idx_payment_retries_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment retry tracking table';

-- Payment retry history table to track individual retry attempts
CREATE TABLE payment_retry_histories (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_retry_id BIGINT UNSIGNED NOT NULL,
    payment_record_id BIGINT UNSIGNED NOT NULL,
    
    -- Attempt Information
    attempt_number INT NOT NULL COMMENT 'Which attempt this was',
    attempted_at TIMESTAMP NULL COMMENT 'When this attempt was made',
    status VARCHAR(20) NOT NULL COMMENT 'success, failed, timeout, error',
    error_message VARCHAR(500) NULL COMMENT 'Error message if failed',
    
    -- Standard timestamp fields
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Indexes
    INDEX idx_payment_retry_histories_payment_retry_id (payment_retry_id),
    INDEX idx_payment_retry_histories_payment_record_id (payment_record_id),
    INDEX idx_payment_retry_histories_attempted_at (attempted_at),
    INDEX idx_payment_retry_histories_status (status),
    INDEX idx_payment_retry_histories_created_at (created_at),
    INDEX idx_payment_retry_histories_deleted_at (deleted_at)
    
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Payment retry attempt history table';