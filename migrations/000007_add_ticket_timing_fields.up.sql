-- Add missing timing and metadata fields to tickets table

-- Add timing fields for response and closure tracking
ALTER TABLE tickets ADD COLUMN first_response_at TIMESTAMP NULL COMMENT 'Time when first admin response was made';
ALTER TABLE tickets ADD COLUMN last_response_at TIMESTAMP NULL COMMENT 'Time when last response was made';
ALTER TABLE tickets ADD COLUMN closed_at TIMESTAMP NULL COMMENT 'Time when ticket was closed';

-- Add metadata fields for tags and custom data
ALTER TABLE tickets ADD COLUMN tags TEXT NULL COMMENT 'Comma-separated tags for ticket categorization';
ALTER TABLE tickets ADD COLUMN metadata JSON NULL COMMENT 'Additional metadata stored as JSON';

-- Add indexes for new timing fields to optimize queries
ALTER TABLE tickets ADD INDEX idx_tickets_first_response_at (first_response_at);
ALTER TABLE tickets ADD INDEX idx_tickets_last_response_at (last_response_at);
ALTER TABLE tickets ADD INDEX idx_tickets_closed_at (closed_at);