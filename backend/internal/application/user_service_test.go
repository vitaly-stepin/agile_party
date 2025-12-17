package application

import (
	"context"
	"errors"
	"testing"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

func TestUserService_JoinRoom_Success(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.JoinRoom(context.Background(), "room123", "user1", "Alice")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_JoinRoom_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.JoinRoom(context.Background(), "", "user1", "Alice")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestUserService_JoinRoom_RoomNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return false, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.JoinRoom(context.Background(), "room123", "user1", "Alice")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestUserService_JoinRoom_InvalidUserName(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.JoinRoom(context.Background(), "room123", "user1", "")

	if err == nil {
		t.Fatal("expected error for empty user name, got nil")
	}
}

func TestUserService_JoinRoom_StateMgrError(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{
		addUserFunc: func(roomID string, user *room.User) error {
			return errors.New("state manager error")
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.JoinRoom(context.Background(), "room123", "user1", "Alice")

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}

func TestUserService_LeaveRoom_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.LeaveRoom(context.Background(), "room123", "user1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_LeaveRoom_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.LeaveRoom(context.Background(), "", "user1")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestUserService_LeaveRoom_EmptyUserID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.LeaveRoom(context.Background(), "room123", "")

	if err != room.ErrInvalidUserID {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestUserService_LeaveRoom_StateMgrError(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		removeUserFunc: func(roomID, userID string) error {
			return errors.New("state manager error")
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.LeaveRoom(context.Background(), "room123", "user1")

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}

func TestUserService_UpdateUserName_Success(t *testing.T) {
	testUser, _ := room.CreateUser("user1", "Alice")

	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		getUserFunc: func(roomID, userID string) (*room.User, error) {
			return testUser, nil
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "room123", "user1", "Bob")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestUserService_UpdateUserName_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "", "user1", "Bob")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestUserService_UpdateUserName_EmptyUserID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "room123", "", "Bob")

	if err != room.ErrInvalidUserID {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestUserService_UpdateUserName_UserNotFound(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		getUserFunc: func(roomID, userID string) (*room.User, error) {
			return nil, room.ErrUserNotFound
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "room123", "user1", "Bob")

	if err != room.ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserService_UpdateUserName_InvalidNewName(t *testing.T) {
	testUser, _ := room.CreateUser("user1", "Alice")

	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		getUserFunc: func(roomID, userID string) (*room.User, error) {
			return testUser, nil
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "room123", "user1", "")

	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestUserService_UpdateUserName_StateMgrError(t *testing.T) {
	testUser, _ := room.CreateUser("user1", "Alice")

	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		getUserFunc: func(roomID, userID string) (*room.User, error) {
			return testUser, nil
		},
		updateUserFunc: func(roomID string, user *room.User) error {
			return errors.New("state manager error")
		},
	}
	service := NewUserService(repo, stateMgr)

	err := service.UpdateUserName(context.Background(), "room123", "user1", "Bob")

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}
