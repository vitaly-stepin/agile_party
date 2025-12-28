package dto

import (
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type TaskResp struct {
	ID          string `json:"id"`
	RoomID      string `json:"roomId"`
	Headline    string `json:"headline"`
	Description string `json:"description,omitempty"`
	TrackerLink string `json:"trackerLink,omitempty"`
	Estimation  string `json:"estimation,omitempty"`
	Position    int    `json:"position"`
}

type CreateTaskReq struct {
	Headline    string `json:"headline"`
	Description string `json:"description,omitempty"`
	TrackerLink string `json:"trackerLink,omitempty"`
}

type UpdateTaskReq struct {
	Headline    string `json:"headline,omitempty"`
	Description string `json:"description,omitempty"`
	TrackerLink string `json:"trackerLink,omitempty"`
}

type ReorderTasksReq struct {
	TaskIDs []string `json:"taskIds"`
}

func FromDomainTask(task *room.Task) *TaskResp {
	if task == nil {
		return nil
	}

	return &TaskResp{
		ID:          task.ID,
		RoomID:      task.RoomID,
		Headline:    task.Headline,
		Description: task.Description,
		TrackerLink: task.TrackerLink,
		Estimation:  task.Estimation,
		Position:    task.Position,
	}
}

func FromDomainTasks(tasks []*room.Task) []*TaskResp {
	if tasks == nil {
		return nil
	}

	result := make([]*TaskResp, len(tasks))
	for i, task := range tasks {
		result[i] = FromDomainTask(task)
	}
	return result
}
