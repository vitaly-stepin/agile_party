package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type TaskRepo struct {
	db *DB
}

func NewTaskRepository(db *DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, task *room.Task) error {
	query := `
        INSERT INTO tasks (id, room_id, headline, description, tracker_link, estimation, position)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `

	_, err := r.db.ExecContext(
		ctx,
		query,
		task.ID,
		task.RoomID,
		task.Headline,
		task.Description,
		task.TrackerLink,
		task.Estimation,
		task.Position,
	)

	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id string) (*room.Task, error) {
	query := `
        SELECT id, room_id, headline, description, tracker_link, estimation, position
        FROM tasks
        WHERE id = $1
    `

	var task room.Task
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&task.ID, &task.RoomID, &task.Headline, &task.Description,
		&task.TrackerLink, &task.Estimation, &task.Position,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, room.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return &task, nil
}

func (r *TaskRepo) GetByRoomID(ctx context.Context, roomID string) ([]*room.Task, error) {
	query := `
        SELECT id, room_id, headline, description, tracker_link, estimation, position
        FROM tasks
        WHERE room_id = $1
        ORDER BY position ASC
    `

	rows, err := r.db.QueryContext(ctx, query, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*room.Task
	for rows.Next() {
		var task room.Task
		err := rows.Scan(
			&task.ID,
			&task.RoomID,
			&task.Headline,
			&task.Description,
			&task.TrackerLink,
			&task.Estimation,
			&task.Position,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, &task)
	}

	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *room.Task) error {
	query := `
        UPDATE tasks
        SET headline = $2, description = $3, tracker_link = $4, estimation = $5, position = $6
        WHERE id = $1
    `

	result, err := r.db.ExecContext(
		ctx,
		query,
		task.ID,
		task.Headline,
		task.Description,
		task.TrackerLink,
		task.Estimation,
		task.Position,
	)

	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return room.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id string) error {
	// First get the task to know its position and room
	task, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	query := `DELETE FROM tasks WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return room.ErrTaskNotFound
	}

	// Reorder remaining tasks to fill the gap
	reorderQuery := `
        UPDATE tasks
        SET position = position - 1
        WHERE room_id = $1 AND position > $2
    `
	_, err = r.db.ExecContext(ctx, reorderQuery, task.RoomID, task.Position)
	if err != nil {
		return fmt.Errorf("failed to reorder tasks after deletion: %w", err)
	}

	return nil
}

func (r *TaskRepo) UpdatePositions(ctx context.Context, tasks []*room.Task) error {
	if len(tasks) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `UPDATE tasks SET position = $1 WHERE id = $2`

	for _, task := range tasks {
		_, err := tx.ExecContext(ctx, query, task.Position, task.ID)
		if err != nil {
			return fmt.Errorf("failed to update task position: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TaskRepo) GetNextUnestimatedTask(ctx context.Context, roomID string) (*room.Task, error) {
	query := `
        SELECT id, room_id, headline, description, tracker_link, estimation, position
        FROM tasks
        WHERE room_id = $1 AND (estimation IS NULL OR estimation = '' OR estimation = '?')
        ORDER BY position ASC
        LIMIT 1
    `

	var task room.Task
	err := r.db.QueryRowContext(ctx, query, roomID).Scan(
		&task.ID,
		&task.RoomID,
		&task.Headline,
		&task.Description,
		&task.TrackerLink,
		&task.Estimation,
		&task.Position,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, room.ErrTaskNotFound
		}
		return nil, fmt.Errorf("failed to get next unestimated task: %w", err)
	}

	return &task, nil
}
