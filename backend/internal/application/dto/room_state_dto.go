package dto

import (
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// RoomStateResponse represents the complete live state of a room
type RoomStateResponse struct {
	RoomID          string         `json:"room_id"`
	RoomName        string         `json:"room_name"`
	Users           []UserResponse `json:"users"`
	Votes           []VoteResponse `json:"votes"`
	IsRevealed      bool           `json:"is_revealed"`
	TaskDescription string         `json:"task_description"`
	Average         *float64       `json:"average,omitempty"`
}

// FromDomainRoomState converts domain room state to DTO
func FromDomainRoomState(state *ports.LiveRoomState) *RoomStateResponse {
	if state == nil {
		return nil
	}

	// Convert users map to slice
	users := make([]UserResponse, 0, len(state.Users))
	for _, user := range state.Users {
		users = append(users, *FromDomainUser(user))
	}

	// Convert votes map to slice
	votes := make([]VoteResponse, 0, len(state.Votes))
	for userID, voteValue := range state.Votes {
		if user, ok := state.Users[userID]; ok {
			votes = append(votes, VoteResponse{
				UserID:   userID,
				UserName: user.Name,
				Value:    voteValue,
			})
		}
	}

	// Calculate average if revealed
	var average *float64
	if state.IsRevealed && len(state.Votes) > 0 {
		estimationSvc := room.NewEstimationService()
		// Assume DBS Fibonacci for now (should come from room settings)
		avg, err := estimationSvc.CalculateAverage(state.Votes, room.DbsFibo)
		if err == nil && avg > 0 {
			average = &avg
		}
	}

	return &RoomStateResponse{
		RoomID:          state.RoomID,
		RoomName:        "", // Will be populated by service layer
		Users:           users,
		Votes:           votes,
		IsRevealed:      state.IsRevealed,
		TaskDescription: state.TaskDescription,
		Average:         average,
	}
}
