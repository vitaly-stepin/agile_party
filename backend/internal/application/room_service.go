package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// RoomService orchestrates room-related use cases
type RoomService struct {
	roomRepo  ports.RoomRepo
	stateMgr  ports.RoomStateManager
}

// NewRoomService creates a new RoomService
func NewRoomService(roomRepo ports.RoomRepo, stateMgr ports.RoomStateManager) *RoomService {
	return &RoomService{
		roomRepo:  roomRepo,
		stateMgr:  stateMgr,
	}
}

// CreateRoom creates a new room and returns its metadata
func (s *RoomService) CreateRoom(ctx context.Context, req *dto.CreateRoomRequest) (*dto.CreateRoomResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Create domain entity
	settings := room.RoomSettings{
		VotingSystem: room.VotingSystem(req.VotingSystem),
		AutoReveal:   req.AutoReveal,
	}

	r, err := room.CreateRoom(req.Name, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	// Persist to database
	if err := s.roomRepo.Create(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to persist room: %w", err)
	}

	// Create in-memory state
	if err := s.stateMgr.CreateRoom(r.ID); err != nil {
		return nil, fmt.Errorf("failed to create room state: %w", err)
	}

	return dto.FromDomainRoomForCreate(r), nil
}

// GetRoom retrieves room metadata by ID
func (s *RoomService) GetRoom(ctx context.Context, roomID string) (*dto.RoomResponse, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}

	r, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return dto.FromDomainRoom(r), nil
}

// GetRoomState retrieves the live state of a room
func (s *RoomService) GetRoomState(ctx context.Context, roomID string) (*dto.RoomStateResponse, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}

	// Check room exists in database
	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room existence: %w", err)
	}
	if !exists {
		return nil, room.ErrRoomNotFound
	}

	// Get live state from memory
	state, err := s.stateMgr.GetRoomState(roomID)
	if err != nil {
		return nil, err
	}

	return dto.FromDomainRoomState(state), nil
}
