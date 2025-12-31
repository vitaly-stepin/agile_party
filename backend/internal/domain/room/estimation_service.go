package room

type EstimationService struct{}

func NewEstimationService() *EstimationService {
	return &EstimationService{}
}

// Calculates the average of votes based on DBS Fibonacci algorithm, excluding "?" votes
func (s *EstimationService) CalculateAverage(votes map[string]string, votingSystem VotingSystem) (float64, error) {
	if len(votes) == 0 {
		return 0, ErrNoVotes
	}

	var votesSum float64
	var totalVotes int

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
			votesSum += floatValue
			totalVotes++
		}
	}

	// If all votes were "?", return 0
	if totalVotes == 0 {
		return 0, nil
	}
	if votesSum == 0 {
		return 0, nil
	}

	average := votesSum / float64(totalVotes)
	switch votingSystem {
	case DbsFibo:
		average = RoundToClosestDbsFiboVote(average)
		return average, nil
	default:
		return average, nil
	}
}

func (s *EstimationService) ValidateAllVotes(votes map[string]string, votingSystem VotingSystem) error {
	for _, voteValue := range votes {
		if err := ValidateVote(voteValue, votingSystem); err != nil {
			return err
		}
	}
	return nil
}
