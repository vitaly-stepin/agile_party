package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type UserService struct {
	roomRepo ports.RoomRepo
	stateMgr ports.RoomStateManager
}

func NewUserService(roomRepo ports.RoomRepo, stateMgr ports.RoomStateManager) *UserService {
	return &UserService{
		roomRepo: roomRepo,
		stateMgr: stateMgr,
	}
}

func (s *UserService) JoinRoom(ctx context.Context, roomID, userID, userName string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}

	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to check room existence: %w", err)
	}
	if !exists {
		return room.ErrRoomNotFound
	}

	// Ensure room exists in memory (lazy initialization after restart)
	if !s.stateMgr.RoomExists(roomID) {
		if err := s.stateMgr.NewRoom(roomID); err != nil {
			return fmt.Errorf("failed to initialize room state: %w", err)
		}
	}

	user, err := room.CreateUser(userID, userName)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if err := s.stateMgr.AddUser(roomID, user); err != nil {
		return fmt.Errorf("failed to add user to room: %w", err)
	}

	return nil
}

func (s *UserService) LeaveRoom(ctx context.Context, roomID, userID string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}
	if userID == "" {
		return room.ErrInvalidUserID
	}

	if err := s.stateMgr.RemoveUser(roomID, userID); err != nil {
		return fmt.Errorf("failed to remove user from room: %w", err)
	}

	return nil
}

func (s *UserService) UpdateUserName(ctx context.Context, roomID, userID, newName string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}
	if userID == "" {
		return room.ErrInvalidUserID
	}

	user, err := s.stateMgr.GetUser(roomID, userID)
	if err != nil {
		return err
	}

	if err := user.UpdateName(newName); err != nil {
		return fmt.Errorf("failed to update user name: %w", err)
	}

	if err := s.stateMgr.UpdateUser(roomID, user); err != nil {
		return fmt.Errorf("failed to update user in state: %w", err)
	}

	return nil
}
