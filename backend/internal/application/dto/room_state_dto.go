package dto

import (
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type RoomStateResp struct {
	RoomID          string     `json:"roomId"`
	RoomName        string     `json:"roomName"`
	Users           []UserResp `json:"users"`
	Votes           []VoteResp `json:"votes"`
	IsRevealed      bool       `json:"isRevealed"`
	TaskDescription string     `json:"taskDescription"`
	Average         *float64   `json:"average,omitempty"`
}

func FromDomainRoomState(state *ports.LiveRoomState) *RoomStateResp {
	if state == nil {
		return nil
	}

	users := make([]UserResp, 0, len(state.Users))
	for _, user := range state.Users {
		users = append(users, *FromDomainUser(user))
	}

	votes := make([]VoteResp, 0, len(state.Votes))
	for userID, voteValue := range state.Votes {
		if user, ok := state.Users[userID]; ok {
			votes = append(votes, VoteResp{
				UserID:   userID,
				UserName: user.Name,
				Value:    voteValue,
			})
		}
	}

	var average *float64
	if state.IsRevealed && len(state.Votes) > 0 {
		estimationSvc := room.NewEstimationService()
		avg, err := estimationSvc.CalculateAverage(state.Votes, room.DbsFibo)
		if err == nil && avg > 0 {
			average = &avg
		}
	}

	return &RoomStateResp{
		RoomID:          state.RoomID,
		RoomName:        "", // Will be populated by service layer
		Users:           users,
		Votes:           votes,
		IsRevealed:      state.IsRevealed,
		TaskDescription: state.TaskDescription,
		Average:         average,
	}
}
