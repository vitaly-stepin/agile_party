package ports

import (
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type LiveRoomState struct {
	RoomID          string
	Users           map[string]*room.User // userID -> User
	Votes           map[string]string     // userID -> vote value
	IsRevealed      bool
	TaskDescription string
}

type RoomStateManager interface {
	CreateRoom(roomID string) error
	GetRoomState(roomID string) (*LiveRoomState, error)
	RoomExists(roomID string) bool
	DeleteRoom(roomID string) error

	AddUser(roomID string, user *room.User) error
	RemoveUser(roomID, userID string) error
	GetUser(roomID, userID string) (*room.User, error)
	UpdateUser(roomID string, user *room.User) error
	GetUserCount(roomID string) (int, error)

	SubmitVote(roomID, userID, voteValue string) error
	RevealVotes(roomID string) error
	ClearVotes(roomID string) error
	UpdateTaskDescription(roomID, description string) error
}