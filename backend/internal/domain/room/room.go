package room

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type RoomSettings struct {
	VotingSystem VotingSystem
	AutoReveal   bool
}

type Room struct {
	ID   string
	Name string
	RoomSettings
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRoom(name string, settings RoomSettings) (*Room, error) {
	if err := ValidateRoomName(name); err != nil {
		return nil, err
	}

	roomID := strings.ReplaceAll(uuid.New().String()[:13], "-", "")[:8]
	now := time.Now()
	return &Room{
		ID:           roomID,
		Name:         strings.TrimSpace(name),
		RoomSettings: settings,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ValidateRoomName(name string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return ErrEmptyRoomName
	}

	if len(trimmed) > 255 {
		return ErrInvalidRoomName
	}

	return nil
}

func (r *Room) UpdateName(name string) error {
	if err := ValidateRoomName(name); err != nil {
		return err
	}
	r.Name = strings.TrimSpace(name)
	r.UpdatedAt = time.Now()
	return nil
}

func (r *Room) UpdateSettings(settings RoomSettings) {
	r.RoomSettings = settings
	r.UpdatedAt = time.Now()
}
