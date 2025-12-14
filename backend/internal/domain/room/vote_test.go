package room

import (
	"testing"
)

func TestValidateDbsFiboVote(t *testing.T) {
	tests := []struct {
		name          string
		voteValue     string
		expectedError bool
	}{
		{"valid zero", "0", false},
		{"valid half", "0.5", false},
		{"valid one", "1", false},
		{"valid two", "2", false},
		{"valid three", "3", false},
		{"valid five", "5", false},
		{"valid eight", "8", false},
		{"valid thirteen", "13", false},
		{"valid twenty", "20", false},
		{"valid forty", "40", false},
		{"valid hundred", "100", false},
		{"valid question mark", "?", false},
		{"invalid negative", "-1", true},
		{"invalid large number", "999", true},
		{"invalid decimal", "1.5", true},
		{"invalid string", "abc", true},
		{"invalid empty", "", true},
		{"invalid four", "4", true},
		{"invalid six", "6", true},
		{"invalid seven", "7", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDbsFiboVote(tt.voteValue)

			if tt.expectedError && err == nil {
				t.Errorf("expected error for value %q but got none", tt.voteValue)
			}

			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error for value %q: %v", tt.voteValue, err)
			}
		})
	}
}

func TestCreateVote(t *testing.T) {
	tests := []struct {
		name          string
		voteValue     string
		votingSystem  VotingSystem
		expectedError bool
	}{
		{"valid vote", "5", DbsFibo, false},
		{"valid question mark", "?", DbsFibo, false},
		{"invalid vote", "999", DbsFibo, true},
		{"unknown voting system", "5", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vote, err := CreateVote(tt.voteValue, tt.votingSystem)

			if tt.expectedError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				if vote != nil {
					t.Errorf("expected nil vote but got %v", vote)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if vote == nil {
				t.Errorf("expected vote but got nil")
				return
			}

			if vote.Value != tt.voteValue {
				t.Errorf("expected value %q, got %q", tt.voteValue, vote.Value)
			}
		})
	}
}

func TestVote_IsNumeric(t *testing.T) {
	tests := []struct {
		name       string
		voteValue  string
		isNumeric  bool
	}{
		{"numeric zero", "0", true},
		{"numeric five", "5", true},
		{"numeric decimal", "0.5", true},
		{"question mark", "?", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vote := &Vote{Value: tt.voteValue}

			if vote.IsNumeric() != tt.isNumeric {
				t.Errorf("expected IsNumeric() to be %v for value %q", tt.isNumeric, tt.voteValue)
			}
		})
	}
}

func TestVote_ToFloat(t *testing.T) {
	tests := []struct {
		name          string
		voteValue     string
		expectedFloat float64
		expectedError bool
	}{
		{"zero", "0", 0.0, false},
		{"half", "0.5", 0.5, false},
		{"five", "5", 5.0, false},
		{"thirteen", "13", 13.0, false},
		{"hundred", "100", 100.0, false},
		{"question mark", "?", 0.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vote := &Vote{Value: tt.voteValue}
			floatValue, err := vote.ToFloat()

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

			if floatValue != tt.expectedFloat {
				t.Errorf("expected %v, got %v", tt.expectedFloat, floatValue)
			}
		})
	}
}
