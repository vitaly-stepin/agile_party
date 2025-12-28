package application

import (
	"context"
	"fmt"

	"github.com/vitaly-stepin/agile_party/internal/application/dto"
	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type TaskService struct {
	taskRepo ports.TaskRepo
	roomRepo ports.RoomRepo
}

func NewTaskService(
	taskRepo ports.TaskRepo,
	roomRepo ports.RoomRepo,
) *TaskService {
	return &TaskService{
		taskRepo: taskRepo,
		roomRepo: roomRepo,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, roomID string, req *dto.CreateTaskReq) (*dto.TaskResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	exists, err := s.roomRepo.Exists(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to check room existence: %w", err)
	}
	if !exists {
		return nil, room.ErrRoomNotFound
	}

	tasks, err := s.taskRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	nextPosition := len(tasks) + 1

	task, err := room.NewTask(roomID, req.Headline, nextPosition)
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	if req.Description != "" {
		task.UpdateDescription(req.Description)
	}
	if req.TrackerLink != "" {
		task.UpdateTrackerLink(req.TrackerLink)
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to persist task: %w", err)
	}

	return dto.FromDomainTask(task), nil
}

func (s *TaskService) GetTask(ctx context.Context, taskID string) (*dto.TaskResp, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}
	return dto.FromDomainTask(task), nil
}

func (s *TaskService) GetRoomTasks(ctx context.Context, roomID string) ([]*dto.TaskResp, error) {
	tasks, err := s.taskRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	return dto.FromDomainTasks(tasks), nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID string, req *dto.UpdateTaskReq) (*dto.TaskResp, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	if req.Headline != "" {
		if err := task.UpdateHeadline(req.Headline); err != nil {
			return nil, err
		}
	}

	if req.Description != "" {
		task.UpdateDescription(req.Description)
	}

	if req.TrackerLink != "" {
		task.UpdateTrackerLink(req.TrackerLink)
	}

	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	return dto.FromDomainTask(task), nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID string) error {
	return s.taskRepo.Delete(ctx, taskID)
}

func (s *TaskService) ReorderTasks(ctx context.Context, roomID string, req *dto.ReorderTasksReq) error {
	if req == nil || len(req.TaskIDs) == 0 {
		return fmt.Errorf("task IDs required")
	}

	tasks, err := s.taskRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return fmt.Errorf("failed to get tasks: %w", err)
	}

	taskMap := make(map[string]*room.Task)
	for _, task := range tasks {
		taskMap[task.ID] = task
	}

	if len(req.TaskIDs) != len(tasks) {
		return fmt.Errorf("task count mismatch: expected %d, got %d", len(tasks), len(req.TaskIDs))
	}

	reorderedTasks := make([]*room.Task, len(req.TaskIDs))
	for i, taskID := range req.TaskIDs {
		task, exists := taskMap[taskID]
		if !exists {
			return fmt.Errorf("task not found: %s", taskID)
		}
		task.Position = i + 1
		reorderedTasks[i] = task
	}

	return s.taskRepo.UpdatePositions(ctx, reorderedTasks)
}

func (s *TaskService) GetNextUnestimatedTask(ctx context.Context, roomID string) (*dto.TaskResp, error) {
	task, err := s.taskRepo.GetNextUnestimatedTask(ctx, roomID)
	if err != nil {
		if err == room.ErrTaskNotFound {
			return nil, nil
		}
		return nil, err
	}
	return dto.FromDomainTask(task), nil
}

func (s *TaskService) SaveEstimation(ctx context.Context, roomID, estimation string) error {
	rm, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return err
	}

	currentTask, err := s.taskRepo.GetNextUnestimatedTask(ctx, roomID)
	if err != nil {
		if err == room.ErrTaskNotFound {
			return nil
		}
		return err
	}

	if err := currentTask.SetEstimation(estimation, rm.VotingSystem); err != nil {
		return fmt.Errorf("failed to set estimation: %w", err)
	}

	if err := s.taskRepo.Update(ctx, currentTask); err != nil {
		return fmt.Errorf("failed to save estimation: %w", err)
	}

	return nil
}

func (s *TaskService) SaveEstimationAndMoveNext(ctx context.Context, roomID, estimation string) (*dto.TaskResp, error) {
	rm, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	currentTask, err := s.taskRepo.GetNextUnestimatedTask(ctx, roomID)
	if err != nil {
		if err == room.ErrTaskNotFound {
			return nil, nil
		}
		return nil, err
	}

	if err := currentTask.SetEstimation(estimation, rm.VotingSystem); err != nil {
		return nil, fmt.Errorf("failed to set estimation: %w", err)
	}

	if err := s.taskRepo.Update(ctx, currentTask); err != nil {
		return nil, fmt.Errorf("failed to save estimation: %w", err)
	}

	nextTask, err := s.taskRepo.GetNextUnestimatedTask(ctx, roomID)
	if err != nil {
		if err == room.ErrTaskNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find next task: %w", err)
	}

	return dto.FromDomainTask(nextTask), nil
}
