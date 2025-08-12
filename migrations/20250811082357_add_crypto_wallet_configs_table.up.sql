-- Crypto wallet configurations table
CREATE TABLE crypto_wallet_configs (
    id INT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    
    -- Network and currency info
    network VARCHAR(50) NOT NULL COMMENT 'Network type: trc, polygon, ethereum',
    currency VARCHAR(20) NOT NULL COMMENT 'Currency: USDT, BTC, ETH', 
    symbol VARCHAR(10) NOT NULL COMMENT 'Currency symbol',
    
    -- Wallet configuration
    wallet_address VARCHAR(255) NOT NULL COMMENT 'Wallet address',
    wallet_name VARCHAR(100) COMMENT 'Wallet name',
    contract_address VARCHAR(255) COMMENT 'Token contract address',
    decimals INT DEFAULT 18 COMMENT 'Token decimals',
    min_confirmations INT DEFAULT 1 COMMENT 'Minimum confirmations',
    
    -- Display settings
    display_name VARCHAR(100) NOT NULL COMMENT 'Display name like TRC-USDT',
    description TEXT COMMENT 'Description',
    icon VARCHAR(255) COMMENT 'Icon URL',
    is_enabled BOOLEAN DEFAULT TRUE COMMENT 'Is enabled',
    sort_order INT DEFAULT 0 COMMENT 'Sort order',
    
    -- Amount limits
    min_amount DECIMAL(20,8) DEFAULT 0.01 COMMENT 'Minimum amount',
    max_amount DECIMAL(20,8) DEFAULT 100000.00 COMMENT 'Maximum amount',
    
    -- Fee configuration
    network_fee DECIMAL(20,8) DEFAULT 0 COMMENT 'Network fee',
    processing_fee DECIMAL(5,4) DEFAULT 0 COMMENT 'Processing fee (%)',
    fixed_fee DECIMAL(10,2) DEFAULT 0 COMMENT 'Fixed fee',
    
    -- API configuration for network monitoring
    api_endpoint VARCHAR(255) COMMENT 'Network API endpoint',
    api_key VARCHAR(255) COMMENT 'API key',
    
    -- Status monitoring
    last_check_at TIMESTAMP NULL COMMENT 'Last check time',
    last_tx_hash VARCHAR(255) COMMENT 'Last transaction hash',
    balance DECIMAL(20,8) DEFAULT 0 COMMENT 'Wallet balance',
    is_active BOOLEAN DEFAULT TRUE COMMENT 'Is active',
    health_status VARCHAR(20) DEFAULT 'healthy' COMMENT 'Health status: healthy, warning, error',
    
    -- Address validation
    address_validated BOOLEAN DEFAULT FALSE COMMENT 'Address validated',
    validated_at TIMESTAMP NULL COMMENT 'Validation time',
    validation_hash VARCHAR(64) COMMENT 'Validation hash',
    
    -- Additional data
    metadata JSON COMMENT 'Additional configuration JSON',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP COMMENT 'Created at',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT 'Updated at',
    deleted_at TIMESTAMP NULL COMMENT 'Deleted at',
    
    -- Indexes
    INDEX idx_network (network) COMMENT 'Network index',
    INDEX idx_currency (currency) COMMENT 'Currency index',
    INDEX idx_network_currency (network, currency) COMMENT 'Network currency composite index',
    INDEX idx_is_enabled (is_enabled) COMMENT 'Enabled status index',
    INDEX idx_is_active (is_active) COMMENT 'Active status index',
    INDEX idx_sort_order (sort_order) COMMENT 'Sort order index',
    INDEX idx_last_check_at (last_check_at) COMMENT 'Last check time index',
    INDEX idx_health_status (health_status) COMMENT 'Health status index',
    INDEX idx_address_validated (address_validated) COMMENT 'Address validation index',
    INDEX idx_created_at (created_at) COMMENT 'Created time index',
    INDEX idx_deleted_at (deleted_at) COMMENT 'Deleted time index',
    
    -- Unique constraints
    UNIQUE KEY unique_wallet_address (wallet_address) COMMENT 'Unique wallet address constraint',
    UNIQUE KEY unique_network_currency_wallet (network, currency, wallet_address) COMMENT 'Unique network currency wallet constraint'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='Crypto wallet configurations table';

-- Sample data for TRC-USDT and Polygon-USDT - Update wallet addresses before deployment
INSERT INTO crypto_wallet_configs (
    network, currency, symbol, wallet_address, wallet_name, 
    contract_address, decimals, min_confirmations, display_name, 
    description, min_amount, max_amount, sort_order, is_enabled, is_active
) VALUES 
(
    'trc', 'USDT', 'USDT', 'TRC_WALLET_ADDRESS_PLACEHOLDER', 'TRC-USDT Wallet',
    'TR7NHqjeKQxGTCIsHNJ7qnHgze1WjawFhX', 6, 1, 'TRC-USDT',
    'TRON network USDT token', 1.00, 50000.00, 1, FALSE, TRUE
),
(
    'polygon', 'USDT', 'USDT', 'POLYGON_WALLET_ADDRESS_PLACEHOLDER', 'Polygon-USDT Wallet', 
    '0xc2132D05D31c914a87C6611C10748AEb04B58e8F', 6, 1, 'Polygon-USDT',
    'Polygon network USDT token', 1.00, 50000.00, 2, FALSE, TRUE
);