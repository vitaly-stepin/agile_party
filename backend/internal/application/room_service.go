package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type RoomService struct {
	roomRepo ports.RoomRepo
	stateMgr ports.RoomStateManager
}

func NewRoomService(roomRepo ports.RoomRepo, stateMgr ports.RoomStateManager) *RoomService {
	return &RoomService{
		roomRepo: roomRepo,
		stateMgr: stateMgr,
	}
}

func (s *RoomService) NewRoom(ctx context.Context, req *dto.NewRoomReq) (*dto.NewRoomResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	settings := room.RoomSettings{
		VotingSystem: room.VotingSystem(req.VotingSystem),
		AutoReveal:   req.AutoReveal,
	}

	r, err := room.NewRoom(req.Name, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}

	if err := s.roomRepo.Create(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to persist room: %w", err)
	}

	if err := s.stateMgr.NewRoom(r.ID); err != nil {
		return nil, fmt.Errorf("failed to create room state: %w", err)
	}

	return dto.FromDomainRoomForCreate(r), nil
}

func (s *RoomService) GetRoom(ctx context.Context, roomID string) (*dto.RoomResp, error) {
	if roomID == "" {
		return nil, room.ErrInvalidRoomID
	}

	r, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return dto.FromDomainRoom(r), nil
}

func (s *RoomService) GetRoomState(ctx context.Context, roomID string) (*dto.RoomStateResp, error) {
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

	state, err := s.stateMgr.GetRoomState(roomID)
	if err != nil {
		return nil, err
	}

	response := dto.FromDomainRoomState(state, r.VotingSystem)
	response.RoomName = r.Name

	return response, nil
}

func (s *RoomService) UpdateTaskDescription(ctx context.Context, roomID, description string) error {
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

	if !s.stateMgr.RoomExists(roomID) {
		if err := s.stateMgr.NewRoom(roomID); err != nil {
			return fmt.Errorf("failed to initialize room state: %w", err)
		}
	}

	if err := s.stateMgr.UpdateTaskDescription(roomID, description); err != nil {
		return fmt.Errorf("failed to update task description: %w", err)
	}

	return nil
}

func (s *RoomService) SetActiveTask(roomID, taskID string) error {
	if roomID == "" {
		return room.ErrInvalidRoomID
	}

	if !s.stateMgr.RoomExists(roomID) {
		if err := s.stateMgr.NewRoom(roomID); err != nil {
			return fmt.Errorf("failed to initialize room state: %w", err)
		}
	}

	if err := s.stateMgr.SetActiveTask(roomID, taskID); err != nil {
		return fmt.Errorf("failed to set active task: %w", err)
	}

	return nil
}

func (s *RoomService) GetActiveTask(roomID string) (string, error) {
	if roomID == "" {
		return "", room.ErrInvalidRoomID
	}

	if !s.stateMgr.RoomExists(roomID) {
		return "", nil
	}

	taskID, err := s.stateMgr.GetActiveTask(roomID)
	if err != nil {
		return "", fmt.Errorf("failed to get active task: %w", err)
	}

	return taskID, nil
}
