package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// RoomRepository implements ports.RoomRepo interface using PostgreSQL
type RoomRepository struct {
	db *DB
}

// NewRoomRepository creates a new PostgreSQL room repository
func NewRoomRepository(db *DB) *RoomRepository {
	return &RoomRepository{db: db}
}

// Create inserts a new room into the database
func (r *RoomRepository) Create(ctx context.Context, rm *room.Room) error {
	query := `
		INSERT INTO rooms (id, name, voting_system, auto_reveal, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		rm.ID,
		rm.Name,
		rm.VotingSystem,
		rm.AutoReveal,
		rm.CreatedAt,
		rm.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}

	return nil
}

// GetByID retrieves a room by its ID
func (r *RoomRepository) GetByID(ctx context.Context, id string) (*room.Room, error) {
	query := `
		SELECT id, name, voting_system, auto_reveal, created_at, updated_at
		FROM rooms
		WHERE id = $1
	`

	var rm room.Room
	var votingSystem string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rm.ID,
		&rm.Name,
		&votingSystem,
		&rm.AutoReveal,
		&rm.CreatedAt,
		&rm.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, room.ErrRoomNotFound
		}
		return nil, fmt.Errorf("failed to get room: %w", err)
	}

	rm.VotingSystem = room.VotingSystem(votingSystem)

	return &rm, nil
}

// Update updates an existing room in the database
func (r *RoomRepository) Update(ctx context.Context, rm *room.Room) error {
	query := `
		UPDATE rooms
		SET name = $2, voting_system = $3, auto_reveal = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		rm.ID,
		rm.Name,
		rm.VotingSystem,
		rm.AutoReveal,
		rm.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update room: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return room.ErrRoomNotFound
	}

	return nil
}

// Delete is not implemented for MVP - rooms persist for historical tracking
func (r *RoomRepository) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("delete not implemented: rooms are persisted for historical tracking")
}

// Exists checks if a room exists in the database
func (r *RoomRepository) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check room existence: %w", err)
	}

	return exists, nil
}
