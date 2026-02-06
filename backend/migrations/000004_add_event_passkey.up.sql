-- Add passkey column to events table for draft room access control
ALTER TABLE events ADD COLUMN passkey VARCHAR(100) NOT NULL;
