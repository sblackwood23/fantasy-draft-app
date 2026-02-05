package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sblackwood23/fantasy-draft-app/internal/models"
)

type PlayerRepository struct {
	pool *pgxpool.Pool
}

func NewPlayerRepository(pool *pgxpool.Pool) *PlayerRepository {
	return &PlayerRepository{pool: pool}
}

func (r *PlayerRepository) GetByID(ctx context.Context, id int) (*models.Player, error) {
	query := `
		SELECT id, first_name, last_name, status, country_code
		FROM players
		WHERE id = $1
	`

	var player models.Player
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&player.ID,
		&player.FirstName,
		&player.LastName,
		&player.Status,
		&player.CountryCode,
	)

	if err != nil {
		return nil, err
	}

	return &player, nil
}

// Retrieves all players
func (r *PlayerRepository) GetAll(ctx context.Context) ([]models.Player, error) {
	query := `
		SELECT id, first_name, last_name, status, country_code
		FROM players
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := []models.Player{}
	for rows.Next() {
		var player models.Player
		err := rows.Scan(
			&player.ID,
			&player.FirstName,
			&player.LastName,
			&player.Status,
			&player.CountryCode,
		)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

// Create new record in players table
func (r *PlayerRepository) Create(ctx context.Context, player *models.Player) error {
	query := `
		INSERT INTO players (first_name, last_name, status, country_code)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	err := r.pool.QueryRow(ctx, query,
		player.FirstName,
		player.LastName,
		player.Status,
		player.CountryCode,
	).Scan(&player.ID)

	return err
}

// Update record in players table
func (r *PlayerRepository) Update(ctx context.Context, player *models.Player) error {
	query := `
		UPDATE players SET first_name=$1, last_name=$2, status=$3, country_code=$4
		WHERE id=$5
	`

	commandTag, err := r.pool.Exec(ctx, query,
		player.FirstName,
		player.LastName,
		player.Status,
		player.CountryCode,
		player.ID,
	)

	if err != nil {
		return err
	}

	if commandTag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

// Delete record from players table
func (r *PlayerRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM players
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
