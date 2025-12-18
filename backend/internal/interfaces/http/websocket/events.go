package websocket

// EventType represents WebSocket event types
type EventType string

// Client-to-Server Events
const (
	EventTypeVote           EventType = "vote"
	EventTypeReveal         EventType = "reveal"
	EventTypeClear          EventType = "clear"
	EventTypeUpdateNickname EventType = "update_nickname"
	EventTypeSetTask        EventType = "set_task"
)

// Server-to-Client Events
const (
	EventTypeRoomState      EventType = "room_state"
	EventTypeUserJoined     EventType = "user_joined"
	EventTypeUserLeft       EventType = "user_left"
	EventTypeVoteSubmitted  EventType = "vote_submitted"
	EventTypeVotesRevealed  EventType = "votes_revealed"
	EventTypeVotesCleared   EventType = "votes_cleared"
	EventTypeUserUpdated    EventType = "user_updated"
	EventTypeError          EventType = "error"
)

// Message represents a WebSocket message
type Message struct {
	Type    EventType   `json:"type"`
	Payload interface{} `json:"payload"`
}

// Client-to-Server Payloads

type VotePayload struct {
	Value string `json:"value"`
}

type UpdateNicknamePayload struct {
	Nickname string `json:"nickname"`
}

type SetTaskPayload struct {
	Description string `json:"description"`
}

// Server-to-Client Payloads

type RoomStatePayload struct {
	RoomID          string        `json:"roomId"`
	RoomName        string        `json:"roomName"`
	Users           []UserPayload `json:"users"`
	Votes           []VoteInfo    `json:"votes"`
	IsRevealed      bool          `json:"isRevealed"`
	TaskDescription string        `json:"taskDescription"`
	Average         *float64      `json:"average,omitempty"`
}

type UserPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IsVoted  bool   `json:"isVoted"`
	IsOnline bool   `json:"isOnline"`
}

type VoteInfo struct {
	UserID   string `json:"userId"`
	Value    string `json:"value"`
	Nickname string `json:"nickname"`
}

type UserJoinedPayload struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
}

type UserLeftPayload struct {
	UserID string `json:"userId"`
}

type VoteSubmittedPayload struct {
	UserID   string `json:"userId"`
	HasVoted bool   `json:"hasVoted"`
}

type VotesRevealedPayload struct {
	Votes   []VoteInfo `json:"votes"`
	Average *float64   `json:"average,omitempty"`
}

type VotesClearedPayload struct{}

type UserUpdatedPayload struct {
	UserID   string `json:"userId"`
	Nickname string `json:"nickname"`
}

type ErrorPayload struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}
