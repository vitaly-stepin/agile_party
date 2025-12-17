package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/ports"

// RoomStateResponse represents the complete live state of a room
type RoomStateResponse struct {
	RoomID     string                  `json:"room_id"`
	Users      map[string]*UserResponse `json:"users"`
	Votes      map[string]string       `json:"votes"`       // userID -> vote value
	IsRevealed bool                    `json:"is_revealed"`
}

// FromDomainRoomState converts domain room state to DTO
func FromDomainRoomState(state *ports.LiveRoomState) *RoomStateResponse {
	if state == nil {
		return nil
	}

	return &RoomStateResponse{
		RoomID:     state.RoomID,
		Users:      FromDomainUsers(state.Users),
		Votes:      state.Votes, // Already map[string]string
		IsRevealed: state.IsRevealed,
	}
}
