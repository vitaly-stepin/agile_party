package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type VotingService struct {
	roomRepo      ports.RoomRepo
	stateMgr      ports.RoomStateManager
	estimationSvc *room.EstimationService
}

func NewVotingService(roomRepo ports.RoomRepo, stateMgr ports.RoomStateManager) *VotingService {
	return &VotingService{
		roomRepo:      roomRepo,
		stateMgr:      stateMgr,
		estimationSvc: room.NewEstimationService(),
	}
}

func (s *VotingService) SubmitVote(ctx context.Context, roomID, userID, voteValue string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}
	if userID == "" {
		return room.ErrInvalidUserID
	}

	r, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Ensure room exists in memory (lazy initialization after restart)
	if !s.stateMgr.RoomExists(roomID) {
		if err := s.stateMgr.NewRoom(roomID); err != nil {
			return fmt.Errorf("failed to initialize room state: %w", err)
		}
	}

	_, err = room.CreateVote(voteValue, r.VotingSystem)
	if err != nil {
		return fmt.Errorf("invalid vote: %w", err)
	}

	if err := s.stateMgr.SubmitVote(roomID, userID, voteValue); err != nil {
		return fmt.Errorf("failed to submit vote: %w", err)
	}

	return nil
}

func (s *VotingService) RevealVotes(ctx context.Context, roomID string) (*dto.RevealVotesResp, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}

	r, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Ensure room exists in memory (lazy initialization after restart)
	if !s.stateMgr.RoomExists(roomID) {
		if err := s.stateMgr.NewRoom(roomID); err != nil {
			return nil, fmt.Errorf("failed to initialize room state: %w", err)
		}
	}

	if err := s.stateMgr.RevealVotes(roomID); err != nil {
		return nil, fmt.Errorf("failed to reveal votes: %w", err)
	}

	state, err := s.stateMgr.GetRoomState(roomID)
	if err != nil {
		return nil, err
	}

	response := &dto.RevealVotesResp{
		Votes: state.Votes, // Already map[string]string
	}

	if len(state.Votes) > 0 {
		avg, err := s.estimationSvc.CalculateAverage(state.Votes, r.VotingSystem)
		// Only set average if no error and average is not 0 when all votes are non-numeric
		// When all votes are "?", CalculateAverage returns 0.0 without error
		// We want to distinguish between "average is 0" and "no numeric votes"
		if err == nil && !(avg == 0.0 && s.hasOnlyNonNumericVotes(state.Votes, r.VotingSystem)) {
			response.Average = &avg
		}
		// If error or all votes non-numeric, Average stays nil
	}

	return response, nil
}

// ClearVotes clears all votes in a room for a new round
func (s *VotingService) ClearVotes(ctx context.Context, roomID string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}

	// Verify room exists
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

	// Clear votes in state
	if err := s.stateMgr.ClearVotes(roomID); err != nil {
		return fmt.Errorf("failed to clear votes: %w", err)
	}

	return nil
}

// hasOnlyNonNumericVotes checks if all votes are non-numeric (e.g., "?")
func (s *VotingService) hasOnlyNonNumericVotes(votes map[string]string, votingSystem room.VotingSystem) bool {
	for _, voteValue := range votes {
		vote, err := room.CreateVote(voteValue, votingSystem)
		if err != nil {
			continue
		}
		if vote.IsNumeric() {
			return false
		}
	}
	return true
}
