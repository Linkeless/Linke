-- Add missing fields to tickets table
ALTER TABLE tickets 
ADD COLUMN first_response_at TIMESTAMP NULL AFTER resolution,
ADD COLUMN last_response_at TIMESTAMP NULL AFTER first_response_at,
ADD COLUMN closed_at TIMESTAMP NULL AFTER last_response_at,
ADD COLUMN tags TEXT AFTER closed_at,
ADD COLUMN metadata JSON AFTER tags;

-- Add missing fields to ticket_messages table
ALTER TABLE ticket_messages
ADD COLUMN attachments JSON AFTER is_internal,
ADD COLUMN metadata JSON AFTER attachments;

-- Add indexes for new fields
CREATE INDEX idx_tickets_first_response ON tickets(first_response_at);
CREATE INDEX idx_tickets_last_response ON tickets(last_response_at);
CREATE INDEX idx_tickets_closed ON tickets(closed_at);