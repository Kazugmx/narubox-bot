package misc

import "github.com/gofiber/fiber/v3"

func NotImplemented(c fiber.Ctx) error {
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "not implemented.",
	})
}
