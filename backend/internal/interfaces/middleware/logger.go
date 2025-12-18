package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

// Logger returns a configured logger middleware
func Logger() fiber.Handler {
	return logger.New(logger.Config{
		Format:     "[${time}] ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: "15:04:05",
		TimeZone:   "Local",
	})
}
