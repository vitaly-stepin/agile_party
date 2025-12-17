package websocket

// EventType represents the type of WebSocket event
type EventType string

const (
	// Client to Server events
	EventTypeVote           EventType = "vote"
	EventTypeReveal         EventType = "reveal"
	EventTypeClear          EventType = "clear"
	EventTypeUpdateNickname EventType = "update_nickname"

	// Server to Client events
	EventTypeRoomState     EventType = "room_state"
	EventTypeUserJoined    EventType = "user_joined"
	EventTypeUserLeft      EventType = "user_left"
	EventTypeVoteSubmitted EventType = "vote_submitted"
	EventTypeVotesRevealed EventType = "votes_revealed"
	EventTypeVotesCleared  EventType = "votes_cleared"
	EventTypeUserUpdated   EventType = "user_updated"
	EventTypeError         EventType = "error"
)

// Event represents a WebSocket message
type Event struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// VotePayload is the payload for vote events (client -> server)
type VotePayload struct {
	Value string `json:"value"`
}

// UpdateNicknamePayload is the payload for nickname update events (client -> server)
type UpdateNicknamePayload struct {
	Nickname string `json:"nickname"`
}

// RoomStatePayload is the initial state sent to client on connection
type RoomStatePayload struct {
	RoomID     string         `json:"roomId"`
	RoomName   string         `json:"roomName"`
	Users      []UserPayload  `json:"users"`
	Votes      []VotePayload2 `json:"votes,omitempty"` // Only included if revealed
	IsRevealed bool           `json:"isRevealed"`
	Average    *float64       `json:"average,omitempty"` // Only if revealed
}

// UserPayload represents user information
type UserPayload struct {
	UserID   string `json:"userId"`
	Name     string `json:"name"`
	IsVoted  bool   `json:"isVoted"`
	IsOnline bool   `json:"isOnline"`
}

// VotePayload2 represents a revealed vote
type VotePayload2 struct {
	UserID string `json:"userId"`
	Value  string `json:"value"`
}

// UserJoinedPayload is sent when a user joins
type UserJoinedPayload struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

// UserLeftPayload is sent when a user leaves
type UserLeftPayload struct {
	UserID string `json:"userId"`
}

// VoteSubmittedPayload is sent when a user submits a vote
type VoteSubmittedPayload struct {
	UserID  string `json:"userId"`
	IsVoted bool   `json:"isVoted"`
}

// VotesRevealedPayload is sent when votes are revealed
type VotesRevealedPayload struct {
	Votes   []VotePayload2 `json:"votes"`
	Average *float64       `json:"average,omitempty"`
}

// UserUpdatedPayload is sent when a user updates their nickname
type UserUpdatedPayload struct {
	UserID string `json:"userId"`
	Name   string `json:"name"`
}

// ErrorPayload is sent when an error occurs
type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
