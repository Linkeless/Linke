-- Create payment_methods table for storing user payment methods securely
-- This table stores tokenized payment information, NOT raw payment data
CREATE TABLE payment_methods (
    id BIGSERIAL PRIMARY KEY,
    
    -- Foreign Keys
    user_id BIGINT NOT NULL,
    
    -- Basic Information
    type VARCHAR(50) NOT NULL, -- card, bank_account, digital_wallet, crypto
    gateway VARCHAR(50) NOT NULL, -- epay, epusdt
    method VARCHAR(50) NOT NULL, -- alipay, wechat, usdt, etc.
    display_name VARCHAR(100) NOT NULL,
    
    -- Tokenized Payment Data (PCI Compliant)
    payment_token VARCHAR(255) NOT NULL, -- Gateway payment token
    gateway_customer_id VARCHAR(255), -- Gateway customer ID
    
    -- Masked Display Information (Safe to show)
    masked_info VARCHAR(100), -- e.g., "**** 1234", "ali***@example.com"
    brand VARCHAR(50), -- e.g., "Visa", "Alipay"
    expiry_month INTEGER, -- For cards only
    expiry_year INTEGER, -- For cards only
    
    -- Status and Configuration
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, inactive, expired, invalid
    
    -- Security and Validation
    last_validated_at TIMESTAMP WITH TIME ZONE,
    validation_hash VARCHAR(64),
    
    -- Billing Information (Optional)
    billing_country VARCHAR(10),
    billing_postcode VARCHAR(20),
    
    -- Gateway-Specific Metadata
    gateway_metadata TEXT, -- JSON metadata from gateway
    
    -- Usage Statistics
    last_used_at TIMESTAMP WITH TIME ZONE,
    successful_uses INTEGER NOT NULL DEFAULT 0,
    failed_uses INTEGER NOT NULL DEFAULT 0,
    
    -- Security Tracking
    created_from_ip VARCHAR(45),
    last_update_ip VARCHAR(45),
    
    -- Timestamp Fields
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes for better performance and constraints
CREATE INDEX idx_payment_methods_user_id ON payment_methods(user_id);
CREATE INDEX idx_payment_methods_type ON payment_methods(type);
CREATE INDEX idx_payment_methods_gateway ON payment_methods(gateway);
CREATE INDEX idx_payment_methods_method ON payment_methods(method);
CREATE INDEX idx_payment_methods_payment_token ON payment_methods(payment_token);
CREATE INDEX idx_payment_methods_is_default ON payment_methods(is_default);
CREATE INDEX idx_payment_methods_is_active ON payment_methods(is_active);
CREATE INDEX idx_payment_methods_status ON payment_methods(status);
CREATE INDEX idx_payment_methods_last_validated_at ON payment_methods(last_validated_at);
CREATE INDEX idx_payment_methods_last_used_at ON payment_methods(last_used_at);
CREATE INDEX idx_payment_methods_created_at ON payment_methods(created_at);
CREATE INDEX idx_payment_methods_deleted_at ON payment_methods(deleted_at);

-- Ensure payment tokens are unique per gateway
CREATE UNIQUE INDEX idx_payment_methods_gateway_token ON payment_methods(gateway, payment_token) WHERE deleted_at IS NULL;

-- Ensure only one default payment method per user per gateway
CREATE UNIQUE INDEX idx_payment_methods_user_gateway_default ON payment_methods(user_id, gateway) 
WHERE is_default = TRUE AND deleted_at IS NULL;

-- Add foreign key constraint (assuming users table exists)
-- ALTER TABLE payment_methods ADD CONSTRAINT fk_payment_methods_user_id 
-- FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

-- Create trigger to update updated_at automatically
CREATE OR REPLACE FUNCTION update_payment_methods_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_payment_methods_updated_at
    BEFORE UPDATE ON payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION update_payment_methods_updated_at();

-- Create function to automatically unset other default methods when setting a new default
CREATE OR REPLACE FUNCTION ensure_single_default_payment_method()
RETURNS TRIGGER AS $$
BEGIN
    -- If the new/updated record is being set as default
    IF NEW.is_default = TRUE AND NEW.deleted_at IS NULL THEN
        -- Unset default for all other payment methods for this user and gateway
        UPDATE payment_methods 
        SET is_default = FALSE, updated_at = NOW()
        WHERE user_id = NEW.user_id 
          AND gateway = NEW.gateway 
          AND id != NEW.id 
          AND is_default = TRUE 
          AND deleted_at IS NULL;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ensure_single_default_payment_method
    AFTER INSERT OR UPDATE ON payment_methods
    FOR EACH ROW
    EXECUTE FUNCTION ensure_single_default_payment_method();

-- Add comments for documentation
COMMENT ON TABLE payment_methods IS 'Stores user payment methods with tokenized data for PCI compliance';
COMMENT ON COLUMN payment_methods.payment_token IS 'Tokenized payment data from gateway - NEVER store raw payment info';
COMMENT ON COLUMN payment_methods.gateway_customer_id IS 'Customer ID in the payment gateway system';
COMMENT ON COLUMN payment_methods.masked_info IS 'Safe masked payment info that can be displayed to users';
COMMENT ON COLUMN payment_methods.validation_hash IS 'Hash for payment method integrity validation';
COMMENT ON COLUMN payment_methods.gateway_metadata IS 'JSON metadata from payment gateway';