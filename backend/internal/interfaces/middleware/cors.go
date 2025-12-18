package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORS returns a configured CORS middleware
func CORS() fiber.Handler {
	return cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173, http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
		AllowCredentials: true,
		MaxAge:           300, // 5 minutes
	})
}
