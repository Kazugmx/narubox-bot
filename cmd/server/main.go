package main

import (
	"log"

	"github.com/gofiber/fiber/v3"
	auth "github.com/kazugmx/narubox-bot/internal/auth"
)

func main() {
	app := fiber.New()

	auth.Route(app)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	log.Fatal(app.Listen(":3000"))
}
