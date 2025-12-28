package ports

import (
	"context"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type TaskRepo interface {
	Create(ctx context.Context, task *room.Task) error
	GetByID(ctx context.Context, id string) (*room.Task, error)
	GetByRoomID(ctx context.Context, roomID string) ([]*room.Task, error)
	Update(ctx context.Context, task *room.Task) error
	Delete(ctx context.Context, id string) error

	UpdatePositions(ctx context.Context, tasks []*room.Task) error
	GetNextUnestimatedTask(ctx context.Context, roomID string) (*room.Task, error)
}
