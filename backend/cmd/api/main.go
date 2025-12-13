package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Agile Party - Scrum Poker v0.1.0",
	})

	// Middleware
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Health check endpoint
	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"message": "Agile Party API is running",
		})
	})

	// Placeholder room endpoint
	app.Post("/api/rooms", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"id": "placeholder",
			"message": "Room creation endpoint - to be implemented",
		})
	})

	// Start server
	port := "8080"
	log.Printf("Starting server on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal(err)
	}
}
