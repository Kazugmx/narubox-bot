package auth

import (
	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	router.Get("/delta", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
}
