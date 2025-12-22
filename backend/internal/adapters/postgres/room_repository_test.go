package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/vitaly-stepin/agile_party/internal/adapters/config"
	"github.com/vitaly-stepin/agile_party/internal/domain/room"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()

	cfg := &config.DatabaseConfig{
		Host:            "postgres",
		Port:            "5432",
		User:            "postgres",
		Password:        "postgres",
		DBName:          "agile_party",
		SSLMode:         "disable",
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	db, err := NewDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func cleanupTestDB(t *testing.T, db *DB, roomID string) {
	t.Helper()
	_, _ = db.Exec("DELETE FROM rooms WHERE id = $1", roomID)
}

func TestRoomRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	rm, err := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	defer cleanupTestDB(t, db, rm.ID)

	// Test Create
	err = repo.Create(ctx, rm)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Verify room was created
	exists, err := repo.Exists(ctx, rm.ID)
	if err != nil {
		t.Fatalf("Failed to check room existence: %v", err)
	}
	if !exists {
		t.Error("Room should exist after creation")
	}
}

func TestRoomRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	originalRoom, err := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   true,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	defer cleanupTestDB(t, db, originalRoom.ID)

	err = repo.Create(ctx, originalRoom)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Test GetByID
	retrievedRoom, err := repo.GetByID(ctx, originalRoom.ID)
	if err != nil {
		t.Fatalf("Failed to get room by ID: %v", err)
	}

	// Verify fields
	if retrievedRoom.ID != originalRoom.ID {
		t.Errorf("Expected ID %s, got %s", originalRoom.ID, retrievedRoom.ID)
	}
	if retrievedRoom.Name != originalRoom.Name {
		t.Errorf("Expected Name %s, got %s", originalRoom.Name, retrievedRoom.Name)
	}
	if retrievedRoom.VotingSystem != originalRoom.VotingSystem {
		t.Errorf("Expected VotingSystem %s, got %s", originalRoom.VotingSystem, retrievedRoom.VotingSystem)
	}
	if retrievedRoom.AutoReveal != originalRoom.AutoReveal {
		t.Errorf("Expected AutoReveal %v, got %v", originalRoom.AutoReveal, retrievedRoom.AutoReveal)
	}
}

func TestRoomRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Test GetByID with non-existent room
	_, err := repo.GetByID(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error when getting non-existent room")
	}
	if err != room.ErrRoomNotFound {
		t.Errorf("Expected ErrRoomNotFound, got %v", err)
	}
}

func TestRoomRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	rm, err := room.NewRoom("Original Name", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	defer cleanupTestDB(t, db, rm.ID)

	err = repo.Create(ctx, rm)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Update room name
	err = rm.UpdateName("Updated Name")
	if err != nil {
		t.Fatalf("Failed to update room name: %v", err)
	}

	// Persist update
	err = repo.Update(ctx, rm)
	if err != nil {
		t.Fatalf("Failed to update room: %v", err)
	}

	// Verify update
	retrievedRoom, err := repo.GetByID(ctx, rm.ID)
	if err != nil {
		t.Fatalf("Failed to get room after update: %v", err)
	}

	if retrievedRoom.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got %s", retrievedRoom.Name)
	}
}

func TestRoomRepository_Update_NotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a room entity but don't persist it
	rm, err := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	// Try to update non-existent room
	err = repo.Update(ctx, rm)
	if err == nil {
		t.Fatal("Expected error when updating non-existent room")
	}
	if err != room.ErrRoomNotFound {
		t.Errorf("Expected ErrRoomNotFound, got %v", err)
	}
}

func TestRoomRepository_Exists(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	rm, err := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	defer cleanupTestDB(t, db, rm.ID)

	// Check non-existent room
	exists, err := repo.Exists(ctx, rm.ID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if exists {
		t.Error("Room should not exist before creation")
	}

	// Create room
	err = repo.Create(ctx, rm)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Check existing room
	exists, err = repo.Exists(ctx, rm.ID)
	if err != nil {
		t.Fatalf("Failed to check existence: %v", err)
	}
	if !exists {
		t.Error("Room should exist after creation")
	}
}

func TestRoomRepository_CreateDuplicate(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewRoomRepository(db)
	ctx := context.Background()

	// Create a test room
	rm, err := room.NewRoom("Test Room", room.RoomSettings{
		VotingSystem: room.DbsFibo,
		AutoReveal:   false,
	})
	if err != nil {
		t.Fatalf("Failed to create room entity: %v", err)
	}

	defer cleanupTestDB(t, db, rm.ID)

	// Create room first time
	err = repo.Create(ctx, rm)
	if err != nil {
		t.Fatalf("Failed to create room: %v", err)
	}

	// Try to create duplicate
	err = repo.Create(ctx, rm)
	if err == nil {
		t.Error("Expected error when creating duplicate room")
	}
}
