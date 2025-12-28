package room

import (
	"strings"
	"testing"
)

func TestNewTask(t *testing.T) {
	tests := []struct {
		name          string
		roomID        string
		headline      string
		position      int
		expectedError bool
	}{
		{"Valid task", "room123", "Implement login", 1, false},
		{"Empty headline", "room123", "", 1, true},
		{"Empty roomID", "", "Task", 1, true},
		{"Invalid position", "room123", "Task", 0, true},
		{"Long headline", "room123", strings.Repeat("a", 256), 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := NewTask(tt.roomID, tt.headline, tt.position)
			if tt.expectedError {
				t.Error(t, err)
			}
			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.expectedError && task == nil {
				t.Error("Expected task but got nil")
			}
		})
	}
}

func TestTaskSetEstimation(t *testing.T) {
	task, _ := NewTask("room123", "Test task", 1)

	err := task.SetEstimation("5", DbsFibo)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if task.Estimation != "5" {
		t.Errorf("Expected estimation '5', got '%s'", task.Estimation)
	}

	err = task.SetEstimation("999", DbsFibo)
	if err == nil {
		t.Error("Expected error for invalid vote value")
	}
}
