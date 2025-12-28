package room

import (
	"strings"

	"github.com/google/uuid"
)

type Task struct {
	ID          string
	RoomID      string
	Headline    string
	Description string
	TrackerLink string
	Estimation  string
	Position    int
}

func NewTask(roomID, headline string, position int) (*Task, error) {
	if err := ValidateTaskHeadline(headline); err != nil {
		return nil, err
	}
	if roomID == "" {
		return nil, ErrInvalidRoomID
	}
	if position < 1 {
		return nil, ErrInvalidTaskPosition
	}

	return &Task{
		ID:          uuid.New().String(),
		RoomID:      roomID,
		Headline:    strings.TrimSpace(headline),
		Description: "",
		TrackerLink: "",
		Estimation:  "",
		Position:    position,
	}, nil
}

func ValidateTaskHeadline(headline string) error {
	trimmed := strings.TrimSpace(headline)
	if trimmed == "" {
		return ErrEmptyTaskHeadline
	}
	if len(trimmed) > 255 {
		return ErrTaskHeadlineTooLong
	}
	return nil
}

func (t *Task) UpdateHeadline(headline string) error {
	if err := ValidateTaskHeadline(headline); err != nil {
		return err
	}
	t.Headline = strings.TrimSpace(headline)
	return nil
}

func (t *Task) UpdateDescription(description string) {
	t.Description = strings.TrimSpace(description)
}

func (t *Task) UpdateTrackerLink(link string) {
	t.TrackerLink = strings.TrimSpace(link)
}

func (t *Task) SetEstimation(value string, votingSystem VotingSystem) error {
	if value != "" {
		vote, err := CreateVote(value, votingSystem)
		if err != nil {
			return err
		}
		t.Estimation = vote.Value
	}
	return nil
}

func (t *Task) IsEstimated() bool {
	return t.Estimation != "" && t.Estimation != "?"
}
