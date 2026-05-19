package Auth

import (
	"log"

	"github.com/gofiber/fiber/v3"
)

func Route(router fiber.Router) {
	log.Println("loaded AuthRoute")

	authRoute := router.Group("/auth")

	authRoute.Get("user", func(c fiber.Ctx) error {
		return c.SendString("fuck you!!")
	})

	authRoute.Post("user", func(c fiber.Ctx) error {
		return c.SendString("get out")
	})
}
