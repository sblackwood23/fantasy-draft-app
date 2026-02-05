-- Rename country column to country_code in players table
ALTER TABLE players RENAME COLUMN country TO country_code;

-- Drop old index and create new one with updated name
DROP INDEX IF EXISTS idx_players_country;
CREATE INDEX idx_players_country_code ON players(country_code);
