package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/vitaly-stepin/agile_party/internal/adapters/config"
	"github.com/vitaly-stepin/agile_party/internal/adapters/memory"
	"github.com/vitaly-stepin/agile_party/internal/adapters/postgres"
	"github.com/vitaly-stepin/agile_party/internal/application"
	"github.com/vitaly-stepin/agile_party/internal/interfaces/http/rest"
	"github.com/vitaly-stepin/agile_party/internal/interfaces/middleware"
	ws "github.com/vitaly-stepin/agile_party/internal/interfaces/http/websocket"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := postgres.NewDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connected successfully")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	log.Println("✅ Database migrations applied")

	// Initialize adapters
	roomRepo := postgres.NewRoomRepository(db)
	stateManager := memory.NewRoomStateManager(memory.CleanupConfig{
		CleanupInterval: cfg.Memory.CleanupInterval,
		RoomTTL:         cfg.Memory.RoomTTL,
	})

	log.Println("✅ Adapters initialized")

	// Initialize application services
	roomService := application.NewRoomService(roomRepo, stateManager)
	userService := application.NewUserService(roomRepo, stateManager)
	votingService := application.NewVotingService(roomRepo, stateManager)

	log.Println("✅ Application services initialized")

	// Initialize WebSocket hub
	hub := ws.NewHub()
	go hub.Run()

	log.Println("✅ WebSocket hub started")

	// Initialize handlers
	roomHandler := rest.NewRoomHandler(roomService, userService, votingService)
	wsHandler := ws.NewHandler(hub, roomService, userService, votingService)

	log.Println("✅ Handlers initialized")

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Agile Party - Scrum Poker v0.1.0",
		ServerHeader:          "Agile Party",
		DisableStartupMessage: false,
		ReadTimeout:           cfg.Server.ReadTimeout,
		WriteTimeout:          cfg.Server.WriteTimeout,
	})

	// Global middleware
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	// Health check endpoint
	app.Get("/api/health", roomHandler.Health)

	// REST API routes
	api := app.Group("/api")

	// Room routes
	api.Post("/rooms", roomHandler.CreateRoom)
	api.Get("/rooms/:id", roomHandler.GetRoom)
	api.Get("/rooms/:id/state", roomHandler.GetRoomState)

	// User routes
	api.Post("/rooms/:id/users", roomHandler.JoinRoom)
	api.Delete("/rooms/:id/users/:userId", roomHandler.LeaveRoom)
	api.Patch("/rooms/:id/users/:userId", roomHandler.UpdateUserName)

	// Voting routes
	api.Post("/rooms/:id/votes", roomHandler.SubmitVote)
	api.Post("/rooms/:id/reveal", roomHandler.RevealVotes)
	api.Post("/rooms/:id/clear", roomHandler.ClearVotes)

	// WebSocket route
	app.Get("/ws/rooms/:id", func(c *fiber.Ctx) error {
		// Check if it's a WebSocket upgrade request
		if websocket.IsWebSocketUpgrade(c) {
			return wsHandler.HandleConnection(c)
		}
		return fiber.ErrUpgradeRequired
	})

	log.Printf("✅ Routes configured")
	log.Printf("🚀 Starting server on port %s", cfg.Server.Port)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + cfg.Server.Port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("⏸️  Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
