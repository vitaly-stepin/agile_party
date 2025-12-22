package dto

import "github.com/vitaly-stepin/agile_party/internal/domain/room"

type SubmitVoteReq struct {
	UserID string `json:"userId"`
	Value  string `json:"value"`
}

type VoteResp struct {
	UserID   string `json:"userId"`
	UserName string `json:"userName"`
	Value    string `json:"value"`
}

type RevealVotesResp struct {
	Votes   map[string]string `json:"votes"`   // userID -> vote value
	Average *float64          `json:"average"` // nil if no numeric votes
}

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
