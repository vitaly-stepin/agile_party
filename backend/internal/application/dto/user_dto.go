package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/room"

type UserResp struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"` // Duplicate for WebSocket compatibility, consider refactoring later
	Name     string `json:"name"`
	IsVoted  bool   `json:"isVoted"`
	IsOnline bool   `json:"isOnline"`
}

func FromDomainUser(u *room.User) *UserResp {
	if u == nil {
		return nil
	}
	return &UserResp{
		ID:       u.ID,
		UserID:   u.ID,
		Name:     u.Name,
		IsVoted:  u.IsVoted,
		IsOnline: true, // Users in state are always online
	}
}

func FromDomainUsers(users map[string]*room.User) map[string]*UserResp {
	if users == nil {
		return nil
	}

	result := make(map[string]*UserResp, len(users))
	for id, user := range users {
		result[id] = FromDomainUser(user)
	}
	return result
}
