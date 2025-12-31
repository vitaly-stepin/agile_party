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
			expectedAvg:   5.0, // avg 6.0 rounds to 5.0 (closer to 5 than to 8)
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
			expectedAvg:   8.0, // avg 6.5 rounds to 8.0 (equidistant, rounds up)
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
			expectedAvg:   1.0, // avg 1.17 rounds to 1.0 (closer to 1 than to 2)
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
			expectedAvg:   100.0, // avg 80.0 rounds to 100.0 (closer to 100 than to 40)
			expectedError: false,
		},
		// User-provided examples
		{
			name: "user example 1: [5,3] -> 5",
			votes: map[string]string{
				"user1": "5",
				"user2": "3",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   5.0, // avg 4.0 equidistant, rounds up to 5
			expectedError: false,
		},
		{
			name: "user example 2: [13,5,5] -> 8",
			votes: map[string]string{
				"user1": "13",
				"user2": "5",
				"user3": "5",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   8.0, // avg 7.67 rounds to 8
			expectedError: false,
		},
		{
			name: "user example 3: [5,1,1] -> 2",
			votes: map[string]string{
				"user1": "5",
				"user2": "1",
				"user3": "1",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   2.0, // avg 2.33 rounds to 2
			expectedError: false,
		},
		// Zero votes bug fix test
		{
			name: "all zero votes",
			votes: map[string]string{
				"user1": "0",
				"user2": "0",
				"user3": "0",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   0.0, // should return 0.0, not be treated as "no numeric votes"
			expectedError: false,
		},
		// Additional edge cases
		{
			name: "mixed zeros and other values",
			votes: map[string]string{
				"user1": "0",
				"user2": "0",
				"user3": "1",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   0.5, // avg 0.33 rounds to 0.5 (closer to 0.5 than to 0)
			expectedError: false,
		},
		{
			name: "equidistant case: [1,2] -> 2",
			votes: map[string]string{
				"user1": "1",
				"user2": "2",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   2.0, // avg 1.5 equidistant, rounds up to 2
			expectedError: false,
		},
		{
			name: "equidistant case: [8,13] -> 13",
			votes: map[string]string{
				"user1": "8",
				"user2": "13",
			},
			votingSystem:  DbsFibo,
			expectedAvg:   13.0, // avg 10.5 equidistant, rounds up to 13
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
