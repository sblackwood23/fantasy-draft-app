-- Add event_id to users table to associate users with specific drafts
ALTER TABLE users ADD COLUMN event_id INTEGER REFERENCES events(id) ON DELETE CASCADE;

-- Add unique constraint: username must be unique per event
ALTER TABLE users ADD CONSTRAINT users_event_username_unique UNIQUE (event_id, username);

-- Create index for efficient lookups by event
CREATE INDEX idx_users_event_id ON users(event_id);
