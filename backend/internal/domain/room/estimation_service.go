package room

import "math"

type EstimationService struct{}

func NewEstimationService() *EstimationService {
	return &EstimationService{}
}

// Calculates the average of votes, excluding "?" votes
func (s *EstimationService) CalculateAverage(votes map[string]string, votingSystem VotingSystem) (float64, error) {
	if len(votes) == 0 {
		return 0, ErrNoVotes
	}

	var votes_sum float64
	var total_votes int

	for _, voteValue := range votes {
		vote, err := CreateVote(voteValue, votingSystem)
		if err != nil {
			return 0, err
		}
		if vote.IsNumeric() {
			floatValue, err := vote.ToFloat()
			if err != nil {
				return 0, err
			}
			votes_sum += floatValue
			total_votes++
		}
	}

	// If all votes were "?", return 0
	if total_votes == 0 {
		return 0, nil
	}

	// Add support for more average calculation strategies later
	average := votes_sum / float64(total_votes)
	average = math.Round(average*100) / 100

	return average, nil
}

func (s *EstimationService) ValidateAllVotes(votes map[string]string, votingSystem VotingSystem) error {
	for _, voteValue := range votes {
		if err := ValidateVote(voteValue, votingSystem); err != nil {
			return err
		}
	}
	return nil
}
