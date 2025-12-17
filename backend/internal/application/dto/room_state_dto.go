package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/ports"

// RoomStateResponse represents the complete live state of a room
type RoomStateResponse struct {
	RoomID     string                   `json:"room_id"`
	RoomName   string                   `json:"room_name"`
	Users      []*UserResponse          `json:"users"`
	Votes      []*VoteResponse          `json:"votes,omitempty"` // Only included if revealed
	IsRevealed bool                     `json:"is_revealed"`
	Average    *float64                 `json:"average,omitempty"` // Only if revealed
}

// FromDomainRoomState converts domain room state to DTO
func FromDomainRoomState(state *ports.LiveRoomState) *RoomStateResponse {
	if state == nil {
		return nil
	}

	// Convert votes map to slice
	var votes []*VoteResponse
	if state.IsRevealed && len(state.Votes) > 0 {
		votes = make([]*VoteResponse, 0, len(state.Votes))
		for userID, value := range state.Votes {
			votes = append(votes, &VoteResponse{
				UserID: userID,
				Value:  value,
			})
		}
	}

	return &RoomStateResponse{
		RoomID:     state.RoomID,
		RoomName:   "", // Will be set by caller if needed
		Users:      FromDomainUsersSlice(state.Users),
		Votes:      votes,
		IsRevealed: state.IsRevealed,
		Average:    nil, // Will be set by caller if needed
	}
}
