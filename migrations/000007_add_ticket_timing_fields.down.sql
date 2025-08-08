-- Remove timing and metadata fields from tickets table

-- Remove indexes first
ALTER TABLE tickets DROP INDEX idx_tickets_closed_at;
ALTER TABLE tickets DROP INDEX idx_tickets_last_response_at; 
ALTER TABLE tickets DROP INDEX idx_tickets_first_response_at;

-- Remove the added columns
ALTER TABLE tickets DROP COLUMN metadata;
ALTER TABLE tickets DROP COLUMN tags;
ALTER TABLE tickets DROP COLUMN closed_at;
ALTER TABLE tickets DROP COLUMN last_response_at;
ALTER TABLE tickets DROP COLUMN first_response_at;