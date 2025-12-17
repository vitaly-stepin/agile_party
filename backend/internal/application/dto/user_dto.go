package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/room"

// UserResponse represents a user in the room
type UserResponse struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`   // Alias for ID for WebSocket compatibility
	Name     string `json:"name"`
	IsVoted  bool   `json:"is_voted"`
	IsOnline bool   `json:"is_online"` // Always true for active users
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
		IsOnline: true, // User is online if they exist in the system
	}
}

// FromDomainUsers converts multiple domain users to DTOs (returns map)
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

// FromDomainUsersSlice converts domain users map to a slice of UserResponse
func FromDomainUsersSlice(users map[string]*room.User) []*UserResponse {
	if users == nil {
		return nil
	}

	result := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, FromDomainUser(user))
	}
	return result
}