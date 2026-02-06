package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/sblackwood23/fantasy-draft-app/internal/models"
)

type EventRepository struct {
	pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{pool: pool}
}

// Retrieves a single event by ID
func (r *EventRepository) GetByID(ctx context.Context, id int) (*models.Event, error) {
	query := `
		SELECT id, name, max_picks_per_team, max_teams_per_player,
		       stipulations, status, passkey, created_at, started_at, completed_at
		FROM events
		WHERE id = $1
	`

	var event models.Event
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&event.ID,
		&event.Name,
		&event.MaxPicksPerTeam,
		&event.MaxTeamsPerPlayer,
		&event.Stipulations,
		&event.Status,
		&event.Passkey,
		&event.CreatedAt,
		&event.StartedAt,
		&event.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &event, nil
}

// Retrieves all events
func (r *EventRepository) GetAll(ctx context.Context) ([]models.Event, error) {
	query := `
		SELECT id, name, max_picks_per_team, max_teams_per_player,
		       stipulations, status, passkey, created_at, started_at, completed_at
		FROM events
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := []models.Event{}
	for rows.Next() {
		var event models.Event
		err := rows.Scan(
			&event.ID,
			&event.Name,
			&event.MaxPicksPerTeam,
			&event.MaxTeamsPerPlayer,
			&event.Stipulations,
			&event.Status,
			&event.Passkey,
			&event.CreatedAt,
			&event.StartedAt,
			&event.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// Create new record in events table
func (r *EventRepository) Create(ctx context.Context, event *models.Event) error {
	query := `
    INSERT INTO events (name, max_picks_per_team, max_teams_per_player, stipulations, status, passkey)
    VALUES ($1, $2, $3, $4, $5, $6)
    RETURNING id, created_at
`
	err := r.pool.QueryRow(ctx, query,
		event.Name,
		event.MaxPicksPerTeam,
		event.MaxTeamsPerPlayer,
		event.Stipulations,
		event.Status,
		event.Passkey,
	).Scan(&event.ID, &event.CreatedAt)

	return err
}

// Update record in events table
func (r *EventRepository) Update(ctx context.Context, event *models.Event) error {
	query := `
		UPDATE events SET name=$1, max_picks_per_team=$2, max_teams_per_player=$3, stipulations=$4, status=$5, passkey=$6
		WHERE id=$7
	`

	commandTag, err := r.pool.Exec(ctx, query,
		event.Name,
		event.MaxPicksPerTeam,
		event.MaxTeamsPerPlayer,
		event.Stipulations,
		event.Status,
		event.Passkey,
		event.ID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete record from events table
func (r *EventRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM events
		WHERE id=$1
	`

	commandTag, err := r.pool.Exec(ctx, query, id)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// GetByPasskey retrieves an event by its passkey
func (r *EventRepository) GetByPasskey(ctx context.Context, passkey string) (*models.Event, error) {
	query := `
		SELECT id, name, max_picks_per_team, max_teams_per_player,
		       stipulations, status, passkey, created_at, started_at, completed_at
		FROM events
		WHERE passkey = $1
	`

	var event models.Event
	err := r.pool.QueryRow(ctx, query, passkey).Scan(
		&event.ID,
		&event.Name,
		&event.MaxPicksPerTeam,
		&event.MaxTeamsPerPlayer,
		&event.Stipulations,
		&event.Status,
		&event.Passkey,
		&event.CreatedAt,
		&event.StartedAt,
		&event.CompletedAt,
	)

	if err != nil {
		return nil, err
	}

	return &event, nil
}

// UpdateStatus updates only the status field and corresponding timestamp
// For "in_progress" status, sets started_at to now
// For "completed" status, sets completed_at to now
func (r *EventRepository) UpdateStatus(ctx context.Context, eventID int, status string) error {
	var query string
	switch status {
	case models.EventStatusInProgress:
		query = `UPDATE events SET status = $1, started_at = NOW() WHERE id = $2`
	case models.EventStatusCompleted:
		query = `UPDATE events SET status = $1, completed_at = NOW() WHERE id = $2`
	default:
		query = `UPDATE events SET status = $1 WHERE id = $2`
	}

	commandTag, err := r.pool.Exec(ctx, query, status, eventID)
	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}
