-- Remove added fields from tickets table
ALTER TABLE tickets
DROP COLUMN IF EXISTS metadata,
DROP COLUMN IF EXISTS tags,
DROP COLUMN IF EXISTS closed_at,
DROP COLUMN IF EXISTS last_response_at,
DROP COLUMN IF EXISTS first_response_at;

-- Remove added fields from ticket_messages table
ALTER TABLE ticket_messages
DROP COLUMN IF EXISTS metadata,
DROP COLUMN IF EXISTS attachments;

-- Drop added indexes
DROP INDEX IF EXISTS idx_tickets_closed ON tickets;
DROP INDEX IF EXISTS idx_tickets_last_response ON tickets;
DROP INDEX IF EXISTS idx_tickets_first_response ON tickets;