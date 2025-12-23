package websocket

// WsEventType represents WebSocket event types
type WsEventType string

// Client Events
const (
	EventTypeVote           WsEventType = "vote"
	EventTypeReveal         WsEventType = "reveal"
	EventTypeClear          WsEventType = "clear"
	EventTypeUpdateNickname WsEventType = "update_nickname"
	EventTypeSetTask        WsEventType = "set_task"
)

// Server Events
const (
	EventTypeRoomState     WsEventType = "room_state"
	EventTypeUserJoined    WsEventType = "user_joined"
	EventTypeUserLeft      WsEventType = "user_left"
	EventTypeVoteSubmitted WsEventType = "vote_submitted"
	EventTypeVotesRevealed WsEventType = "votes_revealed"
	EventTypeVotesCleared  WsEventType = "votes_cleared"
	EventTypeUserUpdated   WsEventType = "user_updated"
	EventTypeError         WsEventType = "error"
)

type WsMessage struct {
	Type    WsEventType `json:"type"`
	Payload interface{} `json:"payload"`
}

type VotePayload struct {
	Value string `json:"value"`
}

type UpdateNicknamePayload struct {
	Nickname string `json:"nickname"`
}

type SetTaskPayload struct {
	Description string `json:"description"`
}

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
