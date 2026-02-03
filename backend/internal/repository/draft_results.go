package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sblackwood23/fantasy-draft-app/internal/models"
)

type DraftResultRepository struct {
	pool *pgxpool.Pool
}

func NewDraftResultRepository(pool *pgxpool.Pool) *DraftResultRepository {
	return &DraftResultRepository{pool: pool}
}

// Create inserts a new draft result (pick) into the database
func (r *DraftResultRepository) Create(ctx context.Context, eventID, userID, playerID, pickNumber, round int) (*models.DraftResult, error) {
	query := `
		INSERT INTO draft_results (event_id, user_id, player_id, pick_number, round)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, event_id, user_id, player_id, pick_number, round, created_at
	`

	var result models.DraftResult
	err := r.pool.QueryRow(ctx, query, eventID, userID, playerID, pickNumber, round).Scan(
		&result.ID,
		&result.EventID,
		&result.UserID,
		&result.PlayerID,
		&result.PickNumber,
		&result.Round,
		&result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetByEvent returns all draft results for a given event
func (r *DraftResultRepository) GetByEvent(ctx context.Context, eventID int) ([]models.DraftResult, error) {
	query := `
		SELECT id, event_id, user_id, player_id, pick_number, round, created_at
		FROM draft_results
		WHERE event_id = $1
		ORDER BY pick_number
	`

	rows, err := r.pool.Query(ctx, query, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.DraftResult
	for rows.Next() {
		var result models.DraftResult
		if err := rows.Scan(
			&result.ID,
			&result.EventID,
			&result.UserID,
			&result.PlayerID,
			&result.PickNumber,
			&result.Round,
			&result.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// GetByEventAndUser returns all draft results for a given event and user
func (r *DraftResultRepository) GetByEventAndUser(ctx context.Context, eventID, userID int) ([]models.DraftResult, error) {
	query := `
		SELECT id, event_id, user_id, player_id, pick_number, round, created_at
		FROM draft_results
		WHERE event_id = $1 AND user_id = $2
		ORDER BY pick_number
	`

	rows, err := r.pool.Query(ctx, query, eventID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []models.DraftResult
	for rows.Next() {
		var result models.DraftResult
		if err := rows.Scan(
			&result.ID,
			&result.EventID,
			&result.UserID,
			&result.PlayerID,
			&result.PickNumber,
			&result.Round,
			&result.CreatedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}
