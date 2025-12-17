package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// UserService handles user-related operations within rooms
type UserService struct {
	roomRepo ports.RoomRepo
	stateMgr ports.RoomStateManager
}

// NewUserService creates a new UserService
func NewUserService(roomRepo ports.RoomRepo, stateMgr ports.RoomStateManager) *UserService {
	return &UserService{
		roomRepo: roomRepo,
		stateMgr: stateMgr,
	}
}

// JoinRoom adds a user to a room
func (s *UserService) JoinRoom(ctx context.Context, roomID, userID, userName string) (*dto.UserResponse, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}

	// Verify room exists in database
	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room existence: %w", err)
	}
	if !exists {
		return nil, room.ErrRoomNotFound
	}

	// Create user entity
	user, err := room.CreateUser(userID, userName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Add user to in-memory state
	if err := s.stateMgr.AddUser(roomID, user); err != nil {
		return nil, fmt.Errorf("failed to add user to room: %w", err)
	}

	return dto.FromDomainUser(user), nil
}

// LeaveRoom removes a user from a room
func (s *UserService) LeaveRoom(ctx context.Context, roomID, userID string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}
	if userID == "" {
		return room.ErrInvalidUserID
	}

	// Remove user from in-memory state
	if err := s.stateMgr.RemoveUser(roomID, userID); err != nil {
		return fmt.Errorf("failed to remove user from room: %w", err)
	}

	return nil
}

// UpdateUserName updates a user's name in the room
func (s *UserService) UpdateUserName(ctx context.Context, roomID, userID, newName string) (*dto.UserResponse, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}
	if userID == "" {
		return nil, room.ErrInvalidUserID
	}

	// Get current user
	user, err := s.stateMgr.GetUser(roomID, userID)
	if err != nil {
		return nil, err
	}

	// Update name
	if err := user.UpdateName(newName); err != nil {
		return nil, fmt.Errorf("failed to update user name: %w", err)
	}

	// Update in memory
	if err := s.stateMgr.UpdateUser(roomID, user); err != nil {
		return nil, fmt.Errorf("failed to update user in state: %w", err)
	}

	return dto.FromDomainUser(user), nil
}
