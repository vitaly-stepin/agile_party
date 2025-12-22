package dto

import (
	"time"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type NewRoomReq struct {
	Name         string `json:"name"`
	VotingSystem string `json:"voting_system"`
	AutoReveal   bool   `json:"auto_reveal"`
}

type NewRoomResp struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	VotingSystem string    `json:"voting_system"`
	AutoReveal   bool      `json:"auto_reveal"`
	CreatedAt    time.Time `json:"created_at"`
}

type UpdateRoomReq struct {
	Name         *string `json:"name,omitempty"`
	VotingSystem *string `json:"votingSystem,omitempty"`
	AutoReveal   *bool   `json:"autoReveal,omitempty"`
}

type RoomResp struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	VotingSystem string    `json:"voting_system"`
	AutoReveal   bool      `json:"auto_reveal"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func FromDomainRoom(r *room.Room) *RoomResp { // consider more self explaining naming
	if r == nil {
		return nil
	}
	return &RoomResp{
		ID:           r.ID,
		Name:         r.Name,
		VotingSystem: string(r.VotingSystem),
		AutoReveal:   r.AutoReveal,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
}

func FromDomainRoomForCreate(r *room.Room) *NewRoomResp { // consider more self explaining naming
	if r == nil {
		return nil
	}
	return &NewRoomResp{
		ID:           r.ID,
		Name:         r.Name,
		VotingSystem: string(r.VotingSystem),
		AutoReveal:   r.AutoReveal,
		CreatedAt:    r.CreatedAt,
	}
}
