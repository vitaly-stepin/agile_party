package room

import "errors"

var (
	ErrRoomNotFound     = errors.New("room not found")
	ErrInvalidRoomID    = errors.New("invalid room ID")
	ErrInvalidRoomName  = errors.New("room name must be between 1 and 255 characters")
	ErrEmptyRoomName    = errors.New("room name cannot be empty")

	ErrInvalidUserID      = errors.New("invalid user ID")
	ErrInvalidUserName    = errors.New("user name must be between 1 and 50 characters")
	ErrEmptyUserName      = errors.New("user name cannot be empty")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists in room")

	ErrInvalidVote         = errors.New("invalid vote value")
	ErrVotingSystemUnknown = errors.New("unknown voting system")
	ErrNoVotes             = errors.New("no votes to calculate")
	ErrVotesNotRevealed    = errors.New("votes have not been revealed yet")
	ErrVotesAlreadyRevealed = errors.New("votes are already revealed")

	ErrRoomAlreadyExists = errors.New("room already exists")
	ErrRoomEmpty         = errors.New("room has no users")
)
