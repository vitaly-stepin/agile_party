package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

type liveRoom struct {
	roomID          string
	users           map[string]*room.User
	votes           map[string]string
	isRevealed      bool
	taskDescription string
	activeTaskID    string
	lastAccess      time.Time
}

type RoomStateManager struct {
	mu    sync.RWMutex
	rooms map[string]*liveRoom // rename to activeRooms + update on first user join
	cfg   CleanupConfig
}

type CleanupConfig struct {
	CleanupInterval time.Duration
	RoomTTL         time.Duration
}

func NewRoomStateManager(cfg CleanupConfig) *RoomStateManager {
	manager := &RoomStateManager{
		rooms: make(map[string]*liveRoom),
		cfg:   cfg,
	}

	go manager.backgroundCleanup()

	return manager
}

func (m *RoomStateManager) NewRoom(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[roomID]; exists {
		return fmt.Errorf("room already exists: %s", roomID)
	}

	m.rooms[roomID] = &liveRoom{
		roomID:          roomID,
		users:           make(map[string]*room.User),
		votes:           make(map[string]string),
		isRevealed:      false,
		taskDescription: "",
		activeTaskID:    "",
		lastAccess:      time.Now(),
	}

	return nil
}

func (m *RoomStateManager) GetRoomState(roomID string) (*ports.LiveRoomState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	r.lastAccess = time.Now()

	// Deep copy to prevent external mutations, check later if we have a better alternatives in GO
	usersCopy := make(map[string]*room.User, len(r.users))
	for id, user := range r.users {
		userCopy := *user
		usersCopy[id] = &userCopy
	}

	votesCopy := make(map[string]string, len(r.votes))
	for id, vote := range r.votes {
		votesCopy[id] = vote
	}

	return &ports.LiveRoomState{
		RoomID:          r.roomID,
		Users:           usersCopy,
		Votes:           votesCopy,
		IsRevealed:      r.isRevealed,
		TaskDescription: r.taskDescription,
		ActiveTaskID:    r.activeTaskID,
	}, nil
}

func (m *RoomStateManager) RoomExists(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.rooms[roomID]
	return exists
}

func (m *RoomStateManager) DeleteRoom(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[roomID]; !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	delete(m.rooms, roomID)
	return nil
}

func (m *RoomStateManager) AddUser(roomID string, user *room.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if _, userExists := r.users[user.ID]; userExists {
		return fmt.Errorf("user already exists in room: %s", user.ID)
	}

	// preserve user vote status if they have already voted (reconnection scenario)
	if _, hasVote := r.votes[user.ID]; hasVote {
		user.IsVoted = true
	}

	r.users[user.ID] = user
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) RemoveUser(roomID, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if _, userExists := r.users[userID]; !userExists {
		return fmt.Errorf("user not found in room: %s", userID)
	}

	delete(r.users, userID)
	delete(r.votes, userID)
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) GetUser(roomID, userID string) (*room.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	user, userExists := r.users[userID]
	if !userExists {
		return nil, fmt.Errorf("user not found in room: %s", userID)
	}

	// Return a copy to prevent external mutations
	userCopy := *user
	return &userCopy, nil
}

func (m *RoomStateManager) UpdateUser(roomID string, user *room.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	if _, userExists := r.users[user.ID]; !userExists {
		return fmt.Errorf("user not found in room: %s", user.ID)
	}

	r.users[user.ID] = user
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) GetUserCount(roomID string) (int, error) {
	m.mu.RLock() // consider later if this lock is necessary
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return 0, fmt.Errorf("room not found: %s", roomID)
	}

	return len(r.users), nil
}

func (m *RoomStateManager) SubmitVote(roomID, userID, voteValue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	user, userExists := r.users[userID]
	if !userExists {
		return fmt.Errorf("user not found in room: %s", userID)
	} // can we refactor this to avoid code duplication?

	r.votes[userID] = voteValue
	user.IsVoted = true
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) RevealVotes(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	r.isRevealed = true
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) ClearVotes(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	r.votes = make(map[string]string)
	r.isRevealed = false
	r.activeTaskID = ""

	for _, user := range r.users {
		user.IsVoted = false
	}

	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) UpdateTaskDescription(roomID, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	r.taskDescription = description
	r.lastAccess = time.Now()

	return nil
}

// backgroundCleanup removes empty rooms periodically
func (m *RoomStateManager) backgroundCleanup() {
	ticker := time.NewTicker(m.cfg.CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		m.cleanup()
	}
}

// cleanup removes empty rooms that are inactive beyond TTL
func (m *RoomStateManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for roomID, r := range m.rooms {
		if len(r.users) == 0 && now.Sub(r.lastAccess) > m.cfg.RoomTTL {
			delete(m.rooms, roomID)
		}
	}
}

func (m *RoomStateManager) SetActiveTask(roomID, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	r.activeTaskID = taskID
	r.lastAccess = time.Now()

	return nil
}

func (m *RoomStateManager) GetActiveTask(roomID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return "", fmt.Errorf("room not found: %s", roomID)
	}

	return r.activeTaskID, nil
}

func (m *RoomStateManager) Stats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalUsers := 0
	for _, r := range m.rooms {
		totalUsers += len(r.users)
	}

	return map[string]interface{}{
		"total_rooms":      len(m.rooms),
		"total_users":      totalUsers,
		"cleanup_interval": m.cfg.CleanupInterval.String(),
		"room_ttl":         m.cfg.RoomTTL.String(),
	}
}
