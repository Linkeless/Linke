-- Remove attachments field from ticket_messages table

-- Remove the added column
ALTER TABLE ticket_messages DROP COLUMN attachments;