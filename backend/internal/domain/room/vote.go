package room

import "strconv"

type VotingSystem string

const (
	DbsFibo   VotingSystem = "dbs_fibo"
	Fibonacci VotingSystem = "fibonacci"
)

type Vote struct {
	Value string
}

func CreateVote(value string, system VotingSystem) (*Vote, error) {
	if err := ValidateVote(value, system); err != nil {
		return nil, err
	}
	return &Vote{Value: value}, nil
}

func ValidateVote(value string, system VotingSystem) error {
	switch system {
	case DbsFibo, Fibonacci:
		return ValidateDbsFiboVote(value)
	default:
		return ErrVotingSystemUnknown
	}
}

func ValidateDbsFiboVote(value string) error {
	validVotes := map[string]bool{
		"?":   true,  // Not voted yet
		"0":   true,
		"0.5": true,
		"1":   true,
		"2":   true,
		"3":   true,
		"5":   true,
		"8":   true,
		"13":  true,
		"20":  true,
		"40":  true,
		"100": true,
	}

	if !validVotes[value] {
		return ErrInvalidVote
	}

	return nil
}

func (v *Vote) IsNumeric() bool {
	return v.Value != "?"
}

func (v *Vote) ToFloat() (float64, error) {
	if !v.IsNumeric() {
		return 0, ErrInvalidVote
	}
	return strconv.ParseFloat(v.Value, 64)
}
