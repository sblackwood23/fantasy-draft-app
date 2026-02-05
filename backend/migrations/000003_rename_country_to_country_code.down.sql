-- Revert country_code back to country in players table
ALTER TABLE players RENAME COLUMN country_code TO country;

-- Drop new index and recreate old one
DROP INDEX IF EXISTS idx_players_country_code;
CREATE INDEX idx_players_country ON players(country);
