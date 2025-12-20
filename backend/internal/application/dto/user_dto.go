package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/room"

// UserResponse represents a user in the room
type UserResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`   // Duplicate for WebSocket compatibility
	Name     string `json:"name"`
	IsVoted  bool   `json:"isVoted"`
	IsOnline bool   `json:"isOnline"` // Always true for users in state
}

// FromDomainUser converts a domain user to a DTO
func FromDomainUser(u *room.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:       u.ID,
		UserID:   u.ID,
		Name:     u.Name,
		IsVoted:  u.IsVoted,
		IsOnline: true, // Users in state are always online
	}
}

// FromDomainUsers converts multiple domain users to DTOs
func FromDomainUsers(users map[string]*room.User) map[string]*UserResponse {
	if users == nil {
		return nil
	}

	result := make(map[string]*UserResponse, len(users))
	for id, user := range users {
		result[id] = FromDomainUser(user)
	}
	return result
}