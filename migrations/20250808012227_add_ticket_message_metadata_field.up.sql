-- Add missing metadata field to ticket_messages table

-- Add metadata field for storing additional message metadata as JSON
ALTER TABLE ticket_messages ADD COLUMN metadata JSON NULL COMMENT 'Additional metadata stored as JSON';