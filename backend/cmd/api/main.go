package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	log.Println("Successfully connected to database")

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	// Initialize adapters (infrastructure layer)
	roomRepo := postgres.NewRoomRepository(db)
	roomStateManager := memory.NewRoomStateManager(memory.CleanupConfig{
		CleanupInterval: cfg.Memory.CleanupInterval,
		RoomTTL:         cfg.Memory.RoomTTL,
	})

	// Initialize application services
	roomService := application.NewRoomService(roomRepo, roomStateManager)
	userService := application.NewUserService(roomRepo, roomStateManager)
	votingService := application.NewVotingService(roomRepo, roomStateManager)

	// Initialize WebSocket hub and start it
	hub := ws.NewHub()
	go hub.Run()

	// Initialize handlers
	restHandler := rest.NewRoomHandler(roomService, userService, votingService)
	wsHandler := ws.NewHandler(hub, roomService, userService, votingService)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "Agile Party - Scrum Poker v0.1.0",
		ErrorHandler: customErrorHandler,
	})

	// Global middleware
	app.Use(middleware.Recovery())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS())

	// REST API routes
	api := app.Group("/api")

	// Health check
	api.Get("/health", restHandler.HealthCheck)

	// Room routes
	api.Post("/rooms", restHandler.CreateRoom)
	api.Get("/rooms/:id", restHandler.GetRoom)
	api.Get("/rooms/:id/state", restHandler.GetRoomState)

	// User routes
	api.Post("/rooms/:id/users", restHandler.JoinRoom)
	api.Delete("/rooms/:id/users/:userId", restHandler.LeaveRoom)
	api.Patch("/rooms/:id/users/:userId", restHandler.UpdateUserName)

	// Voting routes
	api.Post("/rooms/:id/votes", restHandler.SubmitVote)
	api.Post("/rooms/:id/reveal", restHandler.RevealVotes)
	api.Post("/rooms/:id/clear", restHandler.ClearVotes)

	// WebSocket routes
	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws/rooms/:id", websocket.New(wsHandler.HandleConnection))

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := app.ShutdownWithContext(ctx); err != nil {
			log.Printf("Error during shutdown: %v", err)
		}

		log.Println("Server shutdown complete")
	}()

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// customErrorHandler handles errors globally
func customErrorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError

	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	log.Printf("Error: %v", err)

	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
