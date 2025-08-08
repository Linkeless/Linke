-- Add missing attachments field to ticket_messages table

-- Add attachments field for storing file attachments as JSON
ALTER TABLE ticket_messages ADD COLUMN attachments JSON NULL COMMENT 'File attachments stored as JSON array';