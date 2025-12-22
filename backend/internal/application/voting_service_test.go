package application

import (
	"context"
	"errors"
	"testing"

	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

func TestVotingService_SubmitVote_Success(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), testRoom.ID, "user1", "5")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestVotingService_SubmitVote_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), "", "user1", "5")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestVotingService_SubmitVote_EmptyUserID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), "room123", "", "5")

	if err != room.ErrInvalidUserID {
		t.Errorf("expected ErrInvalidUserID, got %v", err)
	}
}

func TestVotingService_SubmitVote_RoomNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return nil, room.ErrRoomNotFound
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), "room123", "user1", "5")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestVotingService_SubmitVote_InvalidVote(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), testRoom.ID, "user1", "99")

	if err == nil {
		t.Fatal("expected error for invalid vote, got nil")
	}
}

func TestVotingService_SubmitVote_StateMgrError(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{
		submitVoteFunc: func(roomID, userID, voteValue string) error {
			return errors.New("state manager error")
		},
	}
	service := NewVotingService(repo, stateMgr)

	err := service.SubmitVote(context.Background(), testRoom.ID, "user1", "5")

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}

func TestVotingService_RevealVotes_Success(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{
		getRoomStateFunc: func(roomID string) (*ports.LiveRoomState, error) {
			return &ports.LiveRoomState{
				RoomID: roomID,
				Users:  make(map[string]*room.User),
				Votes: map[string]string{
					"user1": "5",
					"user2": "8",
				},
				IsRevealed: true,
			}, nil
		},
	}
	service := NewVotingService(repo, stateMgr)

	resp, err := service.RevealVotes(context.Background(), testRoom.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected response, got nil")
	}
	if len(resp.Votes) != 2 {
		t.Errorf("expected 2 votes, got %d", len(resp.Votes))
	}
	if resp.Average == nil {
		t.Fatal("expected average, got nil")
	}
	expectedAvg := 6.5
	if *resp.Average != expectedAvg {
		t.Errorf("expected average %.1f, got %.1f", expectedAvg, *resp.Average)
	}
}

func TestVotingService_RevealVotes_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	_, err := service.RevealVotes(context.Background(), "")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestVotingService_RevealVotes_RoomNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return nil, room.ErrRoomNotFound
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	_, err := service.RevealVotes(context.Background(), "room123")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestVotingService_RevealVotes_NoVotes(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{
		getRoomStateFunc: func(roomID string) (*ports.LiveRoomState, error) {
			return &ports.LiveRoomState{
				RoomID:     roomID,
				Users:      make(map[string]*room.User),
				Votes:      make(map[string]string),
				IsRevealed: true,
			}, nil
		},
	}
	service := NewVotingService(repo, stateMgr)

	resp, err := service.RevealVotes(context.Background(), testRoom.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Average != nil {
		t.Errorf("expected nil average for no votes, got %v", resp.Average)
	}
}

func TestVotingService_RevealVotes_OnlyQuestionMarks(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{
		getRoomStateFunc: func(roomID string) (*ports.LiveRoomState, error) {
			return &ports.LiveRoomState{
				RoomID: roomID,
				Users:  make(map[string]*room.User),
				Votes: map[string]string{
					"user1": "?",
					"user2": "?",
				},
				IsRevealed: true,
			}, nil
		},
	}
	service := NewVotingService(repo, stateMgr)

	resp, err := service.RevealVotes(context.Background(), testRoom.ID)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Average != nil {
		t.Errorf("expected nil average for only question marks, got %v", resp.Average)
	}
}

func TestVotingService_RevealVotes_StateMgrError(t *testing.T) {
	testRoom, _ := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})

	repo := &mockRoomRepo{
		getFunc: func(ctx context.Context, id string) (*room.Room, error) {
			return testRoom, nil
		},
	}
	stateMgr := &mockStateManager{
		revealVotesFunc: func(roomID string) error {
			return errors.New("state manager error")
		},
	}
	service := NewVotingService(repo, stateMgr)

	_, err := service.RevealVotes(context.Background(), testRoom.ID)

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}

func TestVotingService_ClearVotes_Success(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.ClearVotes(context.Background(), "room123")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestVotingService_ClearVotes_EmptyRoomID(t *testing.T) {
	repo := &mockRoomRepo{}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.ClearVotes(context.Background(), "")

	if err != room.ErrInvalidRoomID {
		t.Errorf("expected ErrInvalidRoomID, got %v", err)
	}
}

func TestVotingService_ClearVotes_RoomNotFound(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return false, nil
		},
	}
	stateMgr := &mockStateManager{}
	service := NewVotingService(repo, stateMgr)

	err := service.ClearVotes(context.Background(), "room123")

	if err != room.ErrRoomNotFound {
		t.Errorf("expected ErrRoomNotFound, got %v", err)
	}
}

func TestVotingService_ClearVotes_StateMgrError(t *testing.T) {
	repo := &mockRoomRepo{
		existsFunc: func(ctx context.Context, id string) (bool, error) {
			return true, nil
		},
	}
	stateMgr := &mockStateManager{
		clearVotesFunc: func(roomID string) error {
			return errors.New("state manager error")
		},
	}
	service := NewVotingService(repo, stateMgr)

	err := service.ClearVotes(context.Background(), "room123")

	if err == nil {
		t.Fatal("expected error from state manager, got nil")
	}
}
