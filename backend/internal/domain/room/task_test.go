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
				if err == nil {
					t.Error("expected error but got nil")
				}
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

	// Valid vote value
	err := task.SetEstimation("5", DbsFibo)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if task.Estimation != "5" {
		t.Errorf("Expected estimation '5', got '%s'", task.Estimation)
	}

	// Calculated average (not a valid vote but valid estimation)
	err = task.SetEstimation("3.5", DbsFibo)
	if err != nil {
		t.Errorf("Expected no error for calculated average, got: %v", err)
	}
	if task.Estimation != "3.5" {
		t.Errorf("Expected estimation '3.5', got '%s'", task.Estimation)
	}

	// Any string value should be allowed as estimation
	err = task.SetEstimation("4.2", DbsFibo)
	if err != nil {
		t.Errorf("Expected no error for any estimation value, got: %v", err)
	}
}
