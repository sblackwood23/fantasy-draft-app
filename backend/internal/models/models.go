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
	ID                int          `json:"id"`
	Name              string       `json:"name"`
	MaxPicksPerTeam   int          `json:"maxPicksPerTeam"`
	MaxTeamsPerPlayer int          `json:"maxTeamsPerPlayer"`
	Stipulations      Stipulations `json:"stipulations"`
	Status            string       `json:"status"`
	Passkey           *string      `json:"passkey,omitempty"`
	CreatedAt         time.Time    `json:"createdAt"`
	StartedAt         *time.Time   `json:"startedAt,omitempty"`
	CompletedAt       *time.Time   `json:"completedAt,omitempty"`
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
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
}

// User represents a team/participant in the draft
type User struct {
	ID        int       `json:"id"`
	EventID   int       `json:"eventID"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

// DraftResult represents a pick made during a draft
type DraftResult struct {
	ID         int       `json:"id"`
	EventID    int       `json:"eventID"`
	UserID     int       `json:"userID"`
	PlayerID   int       `json:"playerID"`
	PickNumber int       `json:"pickNumber"`
	Round      int       `json:"round"`
	CreatedAt  time.Time `json:"createdAt"`
}
