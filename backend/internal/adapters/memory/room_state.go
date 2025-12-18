package memory

import (
	"fmt"
	"sync"
	"time"

	"github.com/vitaly-stepin/agile_party/internal/domain/ports"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

// liveRoom holds the live state for a single room
type liveRoom struct {
	roomID          string
	users           map[string]*room.User
	votes           map[string]string
	isRevealed      bool
	taskDescription string
	lastAccess      time.Time
}

// RoomStateManager manages in-memory state for all active rooms
type RoomStateManager struct {
	mu    sync.RWMutex
	rooms map[string]*liveRoom
	cfg   CleanupConfig
}

// CleanupConfig holds configuration for background cleanup
type CleanupConfig struct {
	CleanupInterval time.Duration
	RoomTTL         time.Duration
}

// NewRoomStateManager creates a new in-memory room state manager
func NewRoomStateManager(cfg CleanupConfig) *RoomStateManager {
	manager := &RoomStateManager{
		rooms: make(map[string]*liveRoom),
		cfg:   cfg,
	}

	// Start background cleanup goroutine
	go manager.backgroundCleanup()

	return manager
}

// CreateRoom creates a new live room in memory
func (m *RoomStateManager) CreateRoom(roomID string) error {
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
		lastAccess:      time.Now(),
	}

	return nil
}

// GetRoomState retrieves the current state of a room
func (m *RoomStateManager) GetRoomState(roomID string) (*ports.LiveRoomState, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found: %s", roomID)
	}

	// Update last access
	r.lastAccess = time.Now()

	// Deep copy to prevent external mutations
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
	}, nil
}

// RoomExists checks if a room exists in memory
func (m *RoomStateManager) RoomExists(roomID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.rooms[roomID]
	return exists
}

// DeleteRoom removes a room from memory
func (m *RoomStateManager) DeleteRoom(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[roomID]; !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	delete(m.rooms, roomID)
	return nil
}

// AddUser adds a user to a room (or updates if already exists - for reconnections)
func (m *RoomStateManager) AddUser(roomID string, user *room.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	// Allow re-adding existing users (reconnection scenario)
	// This handles cases where the WebSocket disconnected but the cleanup hasn't run yet
	r.users[user.ID] = user
	r.lastAccess = time.Now()

	return nil
}

// RemoveUser removes a user from a room
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
	delete(r.votes, userID) // Also remove their vote
	r.lastAccess = time.Now()

	return nil
}

// GetUser retrieves a user from a room
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

// UpdateUser updates a user's information in a room
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

// GetUserCount returns the number of users in a room
func (m *RoomStateManager) GetUserCount(roomID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return 0, fmt.Errorf("room not found: %s", roomID)
	}

	return len(r.users), nil
}

// SubmitVote records a user's vote
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
	}

	r.votes[userID] = voteValue
	user.IsVoted = true
	r.lastAccess = time.Now()

	return nil
}

// RevealVotes reveals all votes in a room
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

// ClearVotes clears all votes in a room and resets reveal status
func (m *RoomStateManager) ClearVotes(roomID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	r, exists := m.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found: %s", roomID)
	}

	r.votes = make(map[string]string)
	r.isRevealed = false

	// Reset all users' voting status
	for _, user := range r.users {
		user.IsVoted = false
	}

	r.lastAccess = time.Now()

	return nil
}

// UpdateTaskDescription updates the task description for a room
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

// cleanup removes empty rooms or rooms inactive beyond TTL
func (m *RoomStateManager) cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for roomID, r := range m.rooms {
		// Remove room if:
		// 1. No users connected
		// 2. Last access beyond TTL
		if len(r.users) == 0 && now.Sub(r.lastAccess) > m.cfg.RoomTTL {
			delete(m.rooms, roomID)
		}
	}
}

// Stats returns statistics about the in-memory state
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
