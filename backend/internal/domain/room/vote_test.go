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

func TestRoundToClosestDbsFiboVote(t *testing.T) {
	tests := []struct {
		name     string
		average  float64
		expected float64
	}{
		// User-provided test cases
		{
			name:     "user example: 4.0 equidistant between 3 and 5, rounds up",
			average:  4.0,
			expected: 5.0,
		},
		{
			name:     "user example: 7.67 closer to 8",
			average:  7.67,
			expected: 8.0,
		},
		{
			name:     "user example: 2.33 closer to 2",
			average:  2.33,
			expected: 2.0,
		},

		// Exact matches
		{
			name:     "exactly 0",
			average:  0.0,
			expected: 0.0,
		},
		{
			name:     "exactly 0.5",
			average:  0.5,
			expected: 0.5,
		},
		{
			name:     "exactly 1",
			average:  1.0,
			expected: 1.0,
		},
		{
			name:     "exactly 5",
			average:  5.0,
			expected: 5.0,
		},
		{
			name:     "exactly 13",
			average:  13.0,
			expected: 13.0,
		},
		{
			name:     "exactly 100",
			average:  100.0,
			expected: 100.0,
		},

		// Edge cases - below minimum
		{
			name:     "negative value rounds to 0",
			average:  -5.0,
			expected: 0.0,
		},
		{
			name:     "very small negative rounds to 0",
			average:  -0.1,
			expected: 0.0,
		},

		// Edge cases - above maximum
		{
			name:     "above max (150) rounds to 100",
			average:  150.0,
			expected: 100.0,
		},
		{
			name:     "way above max (1000) rounds to 100",
			average:  1000.0,
			expected: 100.0,
		},

		// Between 0 and 0.5
		{
			name:     "0.2 closer to 0",
			average:  0.2,
			expected: 0.0,
		},
		{
			name:     "0.3 closer to 0.5",
			average:  0.3,
			expected: 0.5,
		},
		{
			name:     "0.25 equidistant, rounds up to 0.5",
			average:  0.25,
			expected: 0.5,
		},

		// Between 0.5 and 1
		{
			name:     "0.6 closer to 0.5",
			average:  0.6,
			expected: 0.5,
		},
		{
			name:     "0.8 closer to 1",
			average:  0.8,
			expected: 1.0,
		},
		{
			name:     "0.75 equidistant, rounds up to 1",
			average:  0.75,
			expected: 1.0,
		},

		// Between 1 and 2
		{
			name:     "1.3 closer to 1",
			average:  1.3,
			expected: 1.0,
		},
		{
			name:     "1.7 closer to 2",
			average:  1.7,
			expected: 2.0,
		},
		{
			name:     "1.5 equidistant, rounds up to 2",
			average:  1.5,
			expected: 2.0,
		},

		// Between 2 and 3
		{
			name:     "2.4 closer to 2",
			average:  2.4,
			expected: 2.0,
		},
		{
			name:     "2.6 closer to 3",
			average:  2.6,
			expected: 3.0,
		},
		{
			name:     "2.5 equidistant, rounds up to 3",
			average:  2.5,
			expected: 3.0,
		},

		// Between 3 and 5
		{
			name:     "3.5 closer to 3",
			average:  3.5,
			expected: 3.0,
		},
		{
			name:     "4.5 closer to 5",
			average:  4.5,
			expected: 5.0,
		},
		{
			name:     "4.0 equidistant, rounds up to 5",
			average:  4.0,
			expected: 5.0,
		},

		// Between 5 and 8
		{
			name:     "6.0 closer to 5",
			average:  6.0,
			expected: 5.0,
		},
		{
			name:     "7.0 closer to 8",
			average:  7.0,
			expected: 8.0,
		},
		{
			name:     "6.5 equidistant, rounds up to 8",
			average:  6.5,
			expected: 8.0,
		},

		// Between 8 and 13
		{
			name:     "10.0 closer to 8",
			average:  10.0,
			expected: 8.0,
		},
		{
			name:     "11.0 closer to 13",
			average:  11.0,
			expected: 13.0,
		},
		{
			name:     "10.5 equidistant, rounds up to 13",
			average:  10.5,
			expected: 13.0,
		},

		// Between 13 and 20
		{
			name:     "15.0 closer to 13",
			average:  15.0,
			expected: 13.0,
		},
		{
			name:     "18.0 closer to 20",
			average:  18.0,
			expected: 20.0,
		},
		{
			name:     "16.5 equidistant, rounds up to 20",
			average:  16.5,
			expected: 20.0,
		},

		// Between 20 and 40
		{
			name:     "25.0 closer to 20",
			average:  25.0,
			expected: 20.0,
		},
		{
			name:     "35.0 closer to 40",
			average:  35.0,
			expected: 40.0,
		},
		{
			name:     "30.0 equidistant, rounds up to 40",
			average:  30.0,
			expected: 40.0,
		},

		// Between 40 and 100
		{
			name:     "60.0 closer to 40",
			average:  60.0,
			expected: 40.0,
		},
		{
			name:     "80.0 closer to 100",
			average:  80.0,
			expected: 100.0,
		},
		{
			name:     "70.0 equidistant, rounds up to 100",
			average:  70.0,
			expected: 100.0,
		},

		// Very close to values
		{
			name:     "8.01 very close to 8",
			average:  8.01,
			expected: 8.0,
		},
		{
			name:     "7.99 very close to 8",
			average:  7.99,
			expected: 8.0,
		},
		{
			name:     "13.001 very close to 13",
			average:  13.001,
			expected: 13.0,
		},
		{
			name:     "12.999 very close to 13",
			average:  12.999,
			expected: 13.0,
		},

		// Decimal precision tests
		{
			name:     "0.49 closer to 0.5",
			average:  0.49,
			expected: 0.5,
		},
		{
			name:     "0.51 closer to 0.5",
			average:  0.51,
			expected: 0.5,
		},
		{
			name:     "2.67 closer to 3",
			average:  2.67,
			expected: 3.0,
		},
		{
			name:     "2.33 closer to 2",
			average:  2.33,
			expected: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RoundToClosestDbsFiboVote(tt.average)
			if result != tt.expected {
				t.Errorf("RoundToClosestDbsFiboVote(%v) = %v, expected %v", tt.average, result, tt.expected)
			}
		})
	}
}

func TestGetDbsFiboVotes(t *testing.T) {
	values := GetDbsFiboVotes()

	// Check length
	expectedLength := 11
	if len(values) != expectedLength {
		t.Errorf("expected %d values, got %d", expectedLength, len(values))
	}

	// Check values are in ascending order
	expected := []float64{0, 0.5, 1, 2, 3, 5, 8, 13, 20, 40, 100}
	for i, val := range values {
		if val != expected[i] {
			t.Errorf("at index %d: expected %v, got %v", i, expected[i], val)
		}
	}

	// Verify ascending order
	for i := 1; i < len(values); i++ {
		if values[i] <= values[i-1] {
			t.Errorf("values not in ascending order at index %d: %v <= %v", i, values[i], values[i-1])
		}
	}
}
