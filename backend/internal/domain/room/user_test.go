package room

import (
	"strings"
	"testing"
)

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name          string
		id            string
		userName      string
		expectedError bool
		checkID       bool
	}{
		{"valid user with new ID", "", "John Doe", false, false},
		{"valid user with existing ID", "user123", "Jane Smith", false, true},
		{"empty name", "", "", true, false},
		{"whitespace only name", "", "   ", true, false},
		{"name too long", "", strings.Repeat("a", 51), true, false},
		{"name exactly 50 chars", "", strings.Repeat("a", 50), false, false},
		{"name with spaces", "", "  John Doe  ", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := CreateUser(tt.id, tt.userName)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if user != nil {
					t.Errorf("expected nil user but got %v", user)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if user == nil {
				t.Errorf("expected user but got nil")
				return
			}

			// Check that ID was generated if not provided
			if tt.id == "" && user.ID == "" {
				t.Errorf("expected generated ID but got empty string")
			}

			// Check that provided ID was used
			if tt.checkID && user.ID != tt.id {
				t.Errorf("expected ID %q, got %q", tt.id, user.ID)
			}

			// Check name is trimmed
			expectedName := strings.TrimSpace(tt.userName)
			if user.Name != expectedName {
				t.Errorf("expected name %q, got %q", expectedName, user.Name)
			}

			// Check default IsVoted value
			if user.IsVoted {
				t.Errorf("expected IsVoted to be false but got true")
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name          string
		userName      string
		expectedError error
	}{
		{"valid name", "John Doe", nil},
		{"empty name", "", ErrEmptyUserName},
		{"whitespace only", "   ", ErrEmptyUserName},
		{"name too long", strings.Repeat("a", 51), ErrInvalidUserName},
		{"name exactly 50 chars", strings.Repeat("a", 50), nil},
		{"name with special chars", "John-Doe_123", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.userName)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestUser_UpdateName(t *testing.T) {
	user, err := CreateUser("", "John Doe")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	tests := []struct {
		name          string
		newName       string
		expectedError bool
	}{
		{"valid update", "Jane Smith", false},
		{"empty name", "", true},
		{"name too long", strings.Repeat("a", 51), true},
		{"whitespace name", "   ", true},
		{"name with leading/trailing spaces", "  Alice  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.UpdateName(tt.newName)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			expectedName := strings.TrimSpace(tt.newName)
			if user.Name != expectedName {
				t.Errorf("expected name %q, got %q", expectedName, user.Name)
			}
		})
	}
}
