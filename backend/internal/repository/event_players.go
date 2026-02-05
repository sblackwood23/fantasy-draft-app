package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sblackwood23/fantasy-draft-app/internal/models"
)

type EventPlayerRepository struct {
	pool *pgxpool.Pool
}

func NewEventPlayerRepository(pool *pgxpool.Pool) *EventPlayerRepository {
	return &EventPlayerRepository{pool: pool}
}

// GetPlayerIDsByEvent returns all player IDs for a given event
func (r *EventPlayerRepository) GetPlayerIDsByEvent(ctx context.Context, eventID int) ([]int, error) {
	query := `SELECT player_id FROM event_players WHERE event_id = $1`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playerIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		playerIDs = append(playerIDs, id)
	}

	return playerIDs, nil
}

// GetPlayersByEvent returns full player objects for a given event
func (r *EventPlayerRepository) GetPlayersByEvent(ctx context.Context, eventID int) ([]models.Player, error) {
	query := `
		SELECT p.id, p.first_name, p.last_name, p.status, p.country_code
		FROM players p
		INNER JOIN event_players ep ON p.id = ep.player_id
		WHERE ep.event_id = $1
	`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []models.Player
	for rows.Next() {
		var player models.Player
		if err := rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Status,
			&player.CountryCode,
		); err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

// AddPlayerToEvent adds a player to an event
func (r *EventPlayerRepository) AddPlayerToEvent(ctx context.Context, eventID, playerID int) error {
	query := `INSERT INTO event_players (event_id, player_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.pool.Exec(ctx, query, eventID, playerID)
	return err
}

// AddPlayersToEvent adds multiple players to an event
func (r *EventPlayerRepository) AddPlayersToEvent(ctx context.Context, eventID int, playerIDs []int) error {
	for _, playerID := range playerIDs {
		if err := r.AddPlayerToEvent(ctx, eventID, playerID); err != nil {
			return err
		}
	}
	return nil
}

// RemovePlayerFromEvent removes a player from an event
func (r *EventPlayerRepository) RemovePlayerFromEvent(ctx context.Context, eventID, playerID int) error {
	query := `DELETE FROM event_players WHERE event_id = $1 AND player_id = $2`
	_, err := r.pool.Exec(ctx, query, eventID, playerID)
	return err
}

// ClearEventPlayers removes all players from an event
func (r *EventPlayerRepository) ClearEventPlayers(ctx context.Context, eventID int) error {
	query := `DELETE FROM event_players WHERE event_id = $1`
	_, err := r.pool.Exec(ctx, query, eventID)
	return err
}
