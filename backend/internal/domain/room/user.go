package room

import (
	"strings"

	"github.com/google/uuid"
)

type User struct {
	ID      string
	Name    string
	IsVoted bool
}


func CreateUser(id, name string) (*User, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	// ID is optional and used for session restoration
	userID := id
	if userID == "" {
		userID = uuid.New().String()
	}

	return &User{
		ID:      userID,
		Name:    strings.TrimSpace(name),
		IsVoted: false,
	}, nil
}

func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)

	if trimmed == "" {
		return ErrEmptyUserName
	}
	if len(trimmed) > 50 {
		return ErrInvalidUserName
	}

	return nil
}

func (u *User) UpdateName(name string) error {
	if err := ValidateName(name); err != nil {
		return err
	}
	u.Name = strings.TrimSpace(name)
	return nil
}