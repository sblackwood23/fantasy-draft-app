package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Event status constants
const (
	EventStatusNotStarted = "not_started"
	EventStatusInProgress = "in_progress"
	EventStatusCompleted  = "completed"
)

// Event represents a draft event with configuration
type Event struct {
	ID                 int              `json:"id"`
	Name               string           `json:"name"`
	MaxPicksPerTeam    int              `json:"max_picks_per_team"`
	MaxTeamsPerPlayer  int              `json:"max_teams_per_player"`
	Stipulations       Stipulations     `json:"stipulations"`
	Status             string           `json:"status"`
	CreatedAt          time.Time        `json:"created_at"`
	StartedAt          *time.Time       `json:"started_at,omitempty"`
	CompletedAt        *time.Time       `json:"completed_at,omitempty"`
}

// Stipulations represents JSONB draft rules stored in events table
type Stipulations map[string]interface{}

// Value implements driver.Valuer for database storage
func (s Stipulations) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database retrieval
func (s *Stipulations) Scan(value interface{}) error {
	if value == nil {
		*s = make(Stipulations)
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Player represents a player in the draft pool
type Player struct {
	ID          int    `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Status      string `json:"status"`
	CountryCode string `json:"country_code"`
}

// User represents a team/participant in the draft
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// DraftResult represents a pick made during a draft
type DraftResult struct {
	ID         int       `json:"id"`
	EventID    int       `json:"event_id"`
	UserID     int       `json:"user_id"`
	PlayerID   int       `json:"player_id"`
	PickNumber int       `json:"pick_number"`
	Round      int       `json:"round"`
	CreatedAt  time.Time `json:"created_at"`
}
