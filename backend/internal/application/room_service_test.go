package application

import (
	"context"
	"errors"
	"testing"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// Mock RoomRepo
type mockRoomRepo struct {
	createFunc func(ctx context.Context, r *room.Room) error
	getFunc    func(ctx context.Context, id string) (*room.Room, error)
	existsFunc func(ctx context.Context, id string) (bool, error)
}

func (m *mockRoomRepo) Create(ctx context.Context, r *room.Room) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, r)
	}
	return nil
}

func (m *mockRoomRepo) GetByID(ctx context.Context, id string) (*room.Room, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, id)
	}
	return nil, room.ErrRoomNotFound
}

func (m *mockRoomRepo) Update(ctx context.Context, r *room.Room) error {
	return nil
}

func (m *mockRoomRepo) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *mockRoomRepo) Exists(ctx context.Context, id string) (bool, error) {
	if m.existsFunc != nil {
		return m.existsFunc(ctx, id)
	}
	return false, nil
}

// Mock RoomStateManager
type mockStateManager struct {
	createRoomFunc   func(roomID string) error
	getRoomStateFunc func(roomID string) (*ports.LiveRoomState, error)
	roomExistsFunc   func(roomID string) bool
	deleteRoomFunc   func(roomID string) error
	addUserFunc      func(roomID string, user *room.User) error
	removeUserFunc   func(roomID, userID string) error
	getUserFunc      func(roomID, userID string) (*room.User, error)
	updateUserFunc   func(roomID string, user *room.User) error
	getUserCountFunc func(roomID string) (int, error)
	submitVoteFunc   func(roomID, userID, voteValue string) error
	revealVotesFunc  func(roomID string) error
	clearVotesFunc   func(roomID string) error
}

func (m *mockStateManager) CreateRoom(roomID string) error {
	if m.createRoomFunc != nil {
		return m.createRoomFunc(roomID)
	}
	return nil
}

func (m *mockStateManager) GetRoomState(roomID string) (*ports.LiveRoomState, error) {
	if m.getRoomStateFunc != nil {
		return m.getRoomStateFunc(roomID)
	}
	return &ports.LiveRoomState{
		RoomID:     roomID,
		Users:      make(map[string]*room.User),
		Votes:      make(map[string]string),
		IsRevealed: false,
	}, nil
}

func (m *mockStateManager) RoomExists(roomID string) bool {
	if m.roomExistsFunc != nil {
		return m.roomExistsFunc(roomID)
	}
	return true
}

func (m *mockStateManager) DeleteRoom(roomID string) error {
	if m.deleteRoomFunc != nil {
		return m.deleteRoomFunc(roomID)
	}
	return nil
}

func (m *mockStateManager) AddUser(roomID string, user *room.User) error {
	if m.addUserFunc != nil {
		return m.addUserFunc(roomID, user)
	}
	return nil
}

func (m *mockStateManager) RemoveUser(roomID, userID string) error {
	if m.removeUserFunc != nil {
		return m.removeUserFunc(roomID, userID)
	}
	return nil
}

func (m *mockStateManager) GetUser(roomID, userID string) (*room.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(roomID, userID)
	}
	return nil, room.ErrUserNotFound
}

func (m *mockStateManager) UpdateUser(roomID string, user *room.User) error {
	if m.updateUserFunc != nil {
		return m.updateUserFunc(roomID, user)
	}
	return nil
}

func (m *mockStateManager) GetUserCount(roomID string) (int, error) {
	if m.getUserCountFunc != nil {
		return m.getUserCountFunc(roomID)
	}
	return 0, nil
}

func (m *mockStateManager) SubmitVote(roomID, userID, voteValue string) error {
	if m.submitVoteFunc != nil {
		return m.submitVoteFunc(roomID, userID, voteValue)
	}
	return nil
}

func (m *mockStateManager) RevealVotes(roomID string) error {
	if m.revealVotesFunc != nil {
		return m.revealVotesFunc(roomID)
	}
	return nil
}

func (m *mockStateManager) ClearVotes(roomID string) error {
	if m.clearVotesFunc != nil {
		return m.clearVotesFunc(roomID)
	}
	return nil
}

// Tests for RoomService

func TestRoomService_CreateRoom_Success(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	req := &dto.CreateRoomRequest{
		Name:         "Sprint Planning",
		VotingSystem: "dbs_fibo",
		AutoReveal:   false,
	}

	resp, err := service.CreateRoom(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.Name != "Sprint Planning" {
		t.Errorf("expected name 'Sprint Planning', got '%s'", resp.Name)
	}
	if resp.VotingSystem != "dbs_fibo" {
		t.Errorf("expected voting system 'dbs_fibo', got '%s'", resp.VotingSystem)
	}
	if len(resp.ID) != 8 {
		t.Errorf("expected ID length 8, got %d", len(resp.ID))
	}
}

func TestRoomService_CreateRoom_NilRequest(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	_, err := service.CreateRoom(context.Background(), nil)

	if err == nil {
		t.Fatal("expected error for nil request, got nil")
	}
}

func TestRoomService_CreateRoom_InvalidName(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	req := &dto.CreateRoomRequest{
		Name:         "",
		VotingSystem: "dbs_fibo",
		AutoReveal:   false,
	}

	_, err := service.CreateRoom(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for empty name, got nil")
	}
}

func TestRoomService_CreateRoom_RepoError(t *testing.T) {
	repo := &mockRoomRepo{
		createFunc: func(ctx context.Context, r *room.Room) error {
			return errors.New("database error")
		},
	}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	req := &dto.CreateRoomRequest{
		Name:         "Sprint Planning",
		VotingSystem: "dbs_fibo",
		AutoReveal:   false,
	}

	_, err := service.CreateRoom(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from repo, got nil")
	}
}

func TestRoomService_CreateRoom_StateMgrError(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{
		createRoomFunc: func(roomID string) error {
			return errors.New("state manager error")
		},
	}
	service := NewRoomService(repo, stateMgr)

	req := &dto.CreateRoomRequest{
		Name:         "Sprint Planning",
		VotingSystem: "dbs_fibo",
		AutoReveal:   false,
	}

	_, err := service.CreateRoom(context.Background(), req)

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}

func TestRoomService_GetRoom_Success(t *testing.T) {
	testRoom, _ := room.CreateRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	resp, err := service.GetRoom(context.Background(), testRoom.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.ID != testRoom.ID {
		t.Errorf("expected ID '%s', got '%s'", testRoom.ID, resp.ID)
	}
}

func TestRoomService_GetRoom_EmptyID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	_, err := service.GetRoom(context.Background(), "")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestRoomService_GetRoom_NotFound(t *testing.T) {
	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return nil, room.ErrRoomNotFound
		},
	}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	_, err := service.GetRoom(context.Background(), "test123")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestRoomService_GetRoomState_Success(t *testing.T) {
	roomID := "test1234"
	testUser, _ := room.CreateUser("user1", "Alice")

	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{
		getRoomStateFunc: func(rID string) (*ports.LiveRoomState, error) {
			return &ports.LiveRoomState{
				RoomID: roomID,
				Users: map[string]*room.User{
					"user1": testUser,
				},
				Votes:      make(map[string]string),
				IsRevealed: false,
			}, nil
		},
	}
	service := NewRoomService(repo, stateMgr)

	resp, err := service.GetRoomState(context.Background(), roomID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if resp.RoomID != roomID {
		t.Errorf("expected room ID '%s', got '%s'", roomID, resp.RoomID)
	}
	if len(resp.Users) != 1 {
		t.Errorf("expected 1 user, got %d", len(resp.Users))
	}
}

func TestRoomService_GetRoomState_EmptyID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	_, err := service.GetRoomState(context.Background(), "")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestRoomService_GetRoomState_RoomNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return false, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewRoomService(repo, stateMgr)

	_, err := service.GetRoomState(context.Background(), "test1234")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}
