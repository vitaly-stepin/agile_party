package middleware

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Recovery returns a configured recovery middleware
func Recovery() fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.Printf("panic: %v\n", e)
		},
	})
}
