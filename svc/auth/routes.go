package Auth

import (
	"context"
	"encoding/json"
	"log"

	"github.com/Kazugmx/narubox-bot/db"
	"github.com/gofiber/fiber/v3"
)

// Queries is a placeholder for the actual Queries type used by Route.
// Define here to avoid undeclared name errors when the real type is
// declared in another package or not needed for build in some contexts.
type Queries struct{}

func Route(router fiber.Router, query *db.Queries, ctx context.Context) {
	log.Println("loaded AuthRoute")

	authRoute := router.Group("/auth")

	authRoute.Post("login", func(c fiber.Ctx) error {
		req := c.Req().Body()
		var login_request UserLoginPayload
		err := json.Unmarshal(req, &login_request)
		if err != nil {
			log.Println("error:", err)
			return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
				"error": "Invalid Request payload.",
			})
		}
		return c.SendString("Hello :3c")
	})
}
