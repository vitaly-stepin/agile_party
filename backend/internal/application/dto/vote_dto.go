package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/room"

// SubmitVoteRequest represents a vote submission request
type SubmitVoteRequest struct {
	Value string `json:"value"`
}

// VoteResponse represents a single vote
type VoteResponse struct {
	UserID string `json:"user_id"`
	Value  string `json:"value"`
}

// RevealVotesResponse represents the result of revealing votes
type RevealVotesResponse struct {
	Votes   map[string]string `json:"votes"`   // userID -> vote value
	Average *float64          `json:"average"` // nil if no numeric votes
}

// FromDomainVotes converts domain votes to DTO
func FromDomainVotes(votes map[string]*room.Vote) map[string]string {
	if votes == nil {
		return nil
	}

	result := make(map[string]string, len(votes))
	for userID, vote := range votes {
		result[userID] = vote.Value
	}
	return result
}