package memory

import (
	"sync"
	"testing"
	"time"

	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

func TestRoomStateManager_CreateRoom(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"

	// Test CreateRoom
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Verify room exists
	if !manager.RoomExists(roomID) {
		t.Error("Room should exist after creation")
	}

	// Test duplicate creation
	err = manager.CreateRoom(roomID)
	if err == nil {
		t.Error("Expected error when creating duplicate room")
	}
}

func TestRoomStateManager_GetRoomState(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test GetRoomState
	state, err := manager.GetRoomState(roomID)
	if err != nil {
		t.Fatalf("Failed to get room state: %v", err)
	}

	if state.RoomID != roomID {
		t.Errorf("Expected room ID %s, got %s", roomID, state.RoomID)
	}
	if len(state.Users) != 0 {
		t.Error("Expected no users initially")
	}
	if len(state.Votes) != 0 {
		t.Error("Expected no votes initially")
	}
	if state.IsRevealed {
		t.Error("Expected votes not revealed initially")
	}
}

func TestRoomStateManager_GetRoomState_NotFound(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	_, err := manager.GetRoomState("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent room state")
	}
}

func TestRoomStateManager_DeleteRoom(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test DeleteRoom
	err = manager.DeleteRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to delete room: %v", err)
	}

	// Verify room no longer exists
	if manager.RoomExists(roomID) {
		t.Error("Room should not exist after deletion")
	}
}

func TestRoomStateManager_AddUser(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test AddUser
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Verify user count
	count, err := manager.GetUserCount(roomID)
	if err != nil {
		t.Fatalf("Failed to get user count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 user, got %d", count)
	}

	// Test duplicate user
	err = manager.AddUser(roomID, user)
	if err == nil {
		t.Error("Expected error when adding duplicate user")
	}
}

func TestRoomStateManager_RemoveUser(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Add a vote for the user
	err = manager.SubmitVote(roomID, user.ID, "5")
	if err != nil {
		t.Fatalf("Failed to submit vote: %v", err)
	}

	// Test RemoveUser
	err = manager.RemoveUser(roomID, user.ID)
	if err != nil {
		t.Fatalf("Failed to remove user: %v", err)
	}

	// Verify user removed
	count, err := manager.GetUserCount(roomID)
	if err != nil {
		t.Fatalf("Failed to get user count: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users, got %d", count)
	}

	// Verify vote also removed
	state, err := manager.GetRoomState(roomID)
	if err != nil {
		t.Fatalf("Failed to get room state: %v", err)
	}
	if _, exists := state.Votes[user.ID]; exists {
		t.Error("User's vote should be removed when user is removed")
	}
}

func TestRoomStateManager_GetUser(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	originalUser, err := room.CreateUser("user1", "Alice")
	err = manager.AddUser(roomID, originalUser)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Test GetUser
	retrievedUser, err := manager.GetUser(roomID, originalUser.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.ID != originalUser.ID {
		t.Errorf("Expected user ID %s, got %s", originalUser.ID, retrievedUser.ID)
	}
	if retrievedUser.Name != originalUser.Name {
		t.Errorf("Expected user name %s, got %s", originalUser.Name, retrievedUser.Name)
	}
}

func TestRoomStateManager_UpdateUser(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Update user name
	err = user.UpdateName("Alice Updated")
	if err != nil {
		t.Fatalf("Failed to update user name: %v", err)
	}

	// Test UpdateUser
	err = manager.UpdateUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify update
	retrievedUser, err := manager.GetUser(roomID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}

	if retrievedUser.Name != "Alice Updated" {
		t.Errorf("Expected updated name 'Alice Updated', got %s", retrievedUser.Name)
	}
}

func TestRoomStateManager_SubmitVote(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Test SubmitVote
	err = manager.SubmitVote(roomID, user.ID, "5")
	if err != nil {
		t.Fatalf("Failed to submit vote: %v", err)
	}

	// Verify vote recorded
	state, err := manager.GetRoomState(roomID)
	if err != nil {
		t.Fatalf("Failed to get room state: %v", err)
	}

	if vote, exists := state.Votes[user.ID]; !exists || vote != "5" {
		t.Errorf("Expected vote '5', got %s (exists: %v)", vote, exists)
	}

	// Verify user marked as voted
	retrievedUser, err := manager.GetUser(roomID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if !retrievedUser.IsVoted {
		t.Error("User should be marked as voted")
	}
}

func TestRoomStateManager_RevealVotes(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test RevealVotes
	err = manager.RevealVotes(roomID)
	if err != nil {
		t.Fatalf("Failed to reveal votes: %v", err)
	}

	// Verify votes revealed
	state, err := manager.GetRoomState(roomID)
	if err != nil {
		t.Fatalf("Failed to get room state: %v", err)
	}

	if !state.IsRevealed {
		t.Error("Votes should be revealed")
	}
}

func TestRoomStateManager_ClearVotes(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	err = manager.SubmitVote(roomID, user.ID, "5")
	if err != nil {
		t.Fatalf("Failed to submit vote: %v", err)
	}

	err = manager.RevealVotes(roomID)
	if err != nil {
		t.Fatalf("Failed to reveal votes: %v", err)
	}

	// Test ClearVotes
	err = manager.ClearVotes(roomID)
	if err != nil {
		t.Fatalf("Failed to clear votes: %v", err)
	}

	// Verify votes cleared
	state, err := manager.GetRoomState(roomID)
	if err != nil {
		t.Fatalf("Failed to get room state: %v", err)
	}

	if len(state.Votes) != 0 {
		t.Error("Votes should be cleared")
	}
	if state.IsRevealed {
		t.Error("Votes should not be revealed after clearing")
	}

	// Verify user voting status reset
	retrievedUser, err := manager.GetUser(roomID, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrievedUser.IsVoted {
		t.Error("User should not be marked as voted after clearing")
	}
}

func TestRoomStateManager_ConcurrentAccess(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Add multiple users concurrently
	var wg sync.WaitGroup
	numUsers := 100

	for i := 0; i < numUsers; i++ {
		wg.Add(1)
		go func(userNum int) {
			defer wg.Done()

			userID := string(rune('A' + (userNum % 26))) + string(rune('0' + (userNum / 26)))
			user, err := room.CreateUser(userID, "User"+userID)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
			err = manager.AddUser(roomID, user)
			if err != nil {
				t.Errorf("Failed to add user %s: %v", userID, err)
			}
		}(i)
	}

	wg.Wait()

	// Verify all users added
	count, err := manager.GetUserCount(roomID)
	if err != nil {
		t.Fatalf("Failed to get user count: %v", err)
	}
	if count != numUsers {
		t.Errorf("Expected %d users, got %d", numUsers, count)
	}
}

func TestRoomStateManager_Cleanup(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 100 * time.Millisecond,
		RoomTTL:         200 * time.Millisecond,
	})

	roomID := "testroom1"
	err := manager.CreateRoom(roomID)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Add a user to prevent immediate cleanup
	user, err := room.CreateUser("user1", "Alice")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	err = manager.AddUser(roomID, user)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	// Remove user to make room empty
	err = manager.RemoveUser(roomID, user.ID)
	if err != nil {
		t.Fatalf("Failed to remove user: %v", err)
	}

	// Wait for cleanup to run
	time.Sleep(500 * time.Millisecond)

	// Verify room was cleaned up
	if manager.RoomExists(roomID) {
		t.Error("Empty room should be cleaned up after TTL")
	}
}

func TestRoomStateManager_Stats(t *testing.T) {
	manager := NewRoomStateManager(CleanupConfig{
		CleanupInterval: 1 * time.Hour,
		RoomTTL:         1 * time.Hour,
	})

	// Create multiple rooms with users
	for i := 0; i < 3; i++ {
		roomID := "room" + string(rune('1'+i))
		err := manager.CreateRoom(roomID)
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		user, err := room.CreateUser("user"+string(rune('1'+i)), "User "+string(rune('1'+i)))
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
		err = manager.AddUser(roomID, user)
		if err != nil {
			t.Fatalf("Failed to add user: %v", err)
		}
	}

	stats := manager.Stats()

	if totalRooms, ok := stats["total_rooms"].(int); !ok || totalRooms != 3 {
		t.Errorf("Expected 3 rooms, got %v", stats["total_rooms"])
	}

	if totalUsers, ok := stats["total_users"].(int); !ok || totalUsers != 3 {
		t.Errorf("Expected 3 users, got %v", stats["total_users"])
	}
}
