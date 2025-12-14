package room

import (
	"testing"
)

func TestEstimationService_CalculateAverage(t *testing.T) {
	service := NewEstimationService()

	tests := []struct {
		name          string
		votes         map[string]string
		votingSystem  VotingSystem
		expectedAvg   float64
		expectedError bool
	}{
		{
			name: "all numeric votes",
			votes: map[string]string{
				"user1": "5",
				"user2": "8",
				"user3": "5",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   6.0,
			expectedError: false,
		},
		{
			name: "votes with question marks",
			votes: map[string]string{
				"user1": "5",
				"user2": "?",
				"user3": "8",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   6.5,
			expectedError: false,
		},
		{
			name: "all question marks",
			votes: map[string]string{
				"user1": "?",
				"user2": "?",
				"user3": "?",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   0,
			expectedError: false,
		},
		{
			name:          "no votes",
			votes:         map[string]string{},
			votingSystem:  DbsFibo,
			expectedAvg:   0,
			expectedError: true,
		},
		{
			name: "votes with decimals",
			votes: map[string]string{
				"user1": "0.5",
				"user2": "1",
				"user3": "2",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   1.17,
			expectedError: false,
		},
		{
			name: "invalid vote value",
			votes: map[string]string{
				"user1": "5",
				"user2": "999", // invalid
			},
			votingSystem:  DbsFibo,
			expectedAvg:   0,
			expectedError: true,
		},
		{
			name: "single vote",
			votes: map[string]string{
				"user1": "13",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   13.0,
			expectedError: false,
		},
		{
			name: "large values",
			votes: map[string]string{
				"user1": "40",
				"user2": "100",
				"user3": "100",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   80.0,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avg, err := service.CalculateAverage(tt.votes, tt.votingSystem)

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

			if avg != tt.expectedAvg {
				t.Errorf("expected average %v, got %v", tt.expectedAvg, avg)
			}
		})
	}
}

func TestEstimationService_ValidateAllVotes(t *testing.T) {
	service := NewEstimationService()

	tests := []struct {
		name          string
		votes         map[string]string
		votingSystem  VotingSystem
		expectedError bool
	}{
		{
			name: "all valid votes",
			votes: map[string]string{
				"user1": "0",
				"user2": "5",
				"user3": "13",
				"user4": "?",
			},
			votingSystem:  DbsFibo,
			expectedError: false,
		},
		{
			name: "one invalid vote",
			votes: map[string]string{
				"user1": "5",
				"user2": "999",
			},
			votingSystem:  DbsFibo,
			expectedError: true,
		},
		{
			name: "empty votes",
			votes: map[string]string{},
			votingSystem: DbsFibo,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateAllVotes(tt.votes, tt.votingSystem)

			if tt.expectedError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectedError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
