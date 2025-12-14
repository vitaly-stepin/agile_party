package ports

import (
	"context"

	"github.com/vitaly-stepin/agile_party/backend/internal/domain/room"
)

type RoomRepo interface {
	Create(ctx context.Context, r *room.Room) error
	GetByID(ctx context.Context, id string) (*room.Room, error)
	Update(ctx context.Context, r *room.Room) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}
