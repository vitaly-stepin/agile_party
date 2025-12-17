package dto

import (
	"time"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// CreateRoomRequest represents a request to create a new room
type CreateRoomRequest struct {
	Name         string `json:"name"`
	VotingSystem string `json:"voting_system"`
	AutoReveal   bool   `json:"auto_reveal"`
}

// CreateRoomResponse represents the response after creating a room
type CreateRoomResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	VotingSystem string    `json:"voting_system"`
	AutoReveal   bool      `json:"auto_reveal"`
	CreatedAt    time.Time `json:"created_at"`
}

// UpdateRoomRequest represents a request to update room settings
type UpdateRoomRequest struct {
	Name         *string `json:"name,omitempty"`
	VotingSystem *string `json:"voting_system,omitempty"`
	AutoReveal   *bool   `json:"auto_reveal,omitempty"`
}

// RoomResponse represents a room with its metadata
type RoomResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	VotingSystem string    `json:"voting_system"`
	AutoReveal   bool      `json:"auto_reveal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// FromDomainRoom converts a domain room to a DTO
func FromDomainRoom(r *room.Room) *RoomResponse {
	if r == nil {
		return nil
	}
	return &RoomResponse{
		ID:           r.ID,
		Name:         r.Name,
		VotingSystem: string(r.VotingSystem),
		AutoReveal:   r.AutoReveal,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

// FromDomainRoomForCreate converts a domain room to a create response DTO
func FromDomainRoomForCreate(r *room.Room) *CreateRoomResponse {
	if r == nil {
		return nil
	}
	return &CreateRoomResponse{
		ID:           r.ID,
		Name:         r.Name,
		VotingSystem: string(r.VotingSystem),
		AutoReveal:   r.AutoReveal,
		CreatedAt:    r.CreatedAt,
	}
}
