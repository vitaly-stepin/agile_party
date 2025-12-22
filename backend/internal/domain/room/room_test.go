package room

import (
	"strings"
	"testing"
)

func TestNewRoom(t *testing.T) {
	tests := []struct {
		name          string
		roomName      string
		settings      RoomSettings
		expectedError bool
	}{
		{
			name:     "valid room with settings",
			roomName: "Sprint Planning",
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: false,
		},
		{
			name:     "valid room with auto reveal",
			roomName: "Daily Standup",
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   true,
			},
			expectedError: false,
		},
		{
			name:     "empty room name",
			roomName: "",
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: true,
		},
		{
			name:     "whitespace only room name",
			roomName: "   ",
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: true,
		},
		{
			name:     "room name too long",
			roomName: strings.Repeat("a", 256),
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: true,
		},
		{
			name:     "room name exactly 255 chars",
			roomName: strings.Repeat("a", 255),
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: false,
		},
		{
			name:     "room name with leading/trailing spaces",
			roomName: "  My Room  ",
			settings: RoomSettings{
				VotingSystem: DbsFibo,
				AutoReveal:   false,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			room, err := NewRoom(tt.roomName, tt.settings)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if room != nil {
					t.Errorf("expected nil room but got %v", room)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if room == nil {
				t.Errorf("expected room but got nil")
				return
			}

			// Check ID was generated and has expected length
			if len(room.ID) != 8 {
				t.Errorf("expected ID length 8, got %d", len(room.ID))
			}

			// Check name is trimmed
			expectedName := strings.TrimSpace(tt.roomName)
			if room.Name != expectedName {
				t.Errorf("expected name %q, got %q", expectedName, room.Name)
			}

			// Check settings are applied
			if room.VotingSystem != tt.settings.VotingSystem {
				t.Errorf("expected voting system %v, got %v", tt.settings.VotingSystem, room.VotingSystem)
			}

			if room.AutoReveal != tt.settings.AutoReveal {
				t.Errorf("expected auto reveal %v, got %v", tt.settings.AutoReveal, room.AutoReveal)
			}

			// Check timestamps
			if room.CreatedAt.IsZero() {
				t.Errorf("expected CreatedAt to be set")
			}

			if room.UpdatedAt.IsZero() {
				t.Errorf("expected UpdatedAt to be set")
			}

			if !room.CreatedAt.Equal(room.UpdatedAt) {
				t.Errorf("expected CreatedAt and UpdatedAt to be equal on creation")
			}
		})
	}
}

func TestValidateRoomName(t *testing.T) {
	tests := []struct {
		name          string
		roomName      string
		expectedError error
	}{
		{"valid name", "Sprint Planning", nil},
		{"empty name", "", ErrEmptyRoomName},
		{"whitespace only", "   ", ErrEmptyRoomName},
		{"name too long", strings.Repeat("a", 256), ErrInvalidRoomName},
		{"name exactly 255 chars", strings.Repeat("a", 255), nil},
		{"name with special chars", "Sprint-Planning_2024", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoomName(tt.roomName)

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

func TestRoom_UpdateName(t *testing.T) {
	room, err := NewRoom("Original Name", RoomSettings{
		VotingSystem: DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	originalUpdatedAt := room.UpdatedAt

	tests := []struct {
		name          string
		newName       string
		expectedError bool
	}{
		{"valid update", "New Name", false},
		{"empty name", "", true},
		{"name too long", strings.Repeat("a", 256), true},
		{"whitespace name", "   ", true},
		{"name with leading/trailing spaces", "  Updated Room  ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := room.UpdateName(tt.newName)

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
			if room.Name != expectedName {
				t.Errorf("expected name %q, got %q", expectedName, room.Name)
			}

			// Check that UpdatedAt was updated
			if !room.UpdatedAt.After(originalUpdatedAt) {
				t.Errorf("expected UpdatedAt to be updated")
			}
		})
	}
}

func TestRoom_UpdateSettings(t *testing.T) {
	room, err := NewRoom("Test Room", RoomSettings{
		VotingSystem: DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	originalUpdatedAt := room.UpdatedAt

	newSettings := RoomSettings{
		VotingSystem: DbsFibo,
		AutoReveal:   true,
	}

	room.UpdateSettings(newSettings)

	if room.VotingSystem != newSettings.VotingSystem {
		t.Errorf("expected voting system %v, got %v", newSettings.VotingSystem, room.VotingSystem)
	}

	if room.AutoReveal != newSettings.AutoReveal {
		t.Errorf("expected auto reveal %v, got %v", newSettings.AutoReveal, room.AutoReveal)
	}

	// Check that UpdatedAt was updated
	if !room.UpdatedAt.After(originalUpdatedAt) {
		t.Errorf("expected UpdatedAt to be updated")
	}
}
