package Auth

import (
	"context"
	"log"

	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/internal/auth"
	"github.com/gofiber/fiber/v3"
)

type Queries struct{}

func Route(
	router fiber.Router,
	query *db.Queries,
	ctx context.Context,
	jwtEngine *jwtOperator.JWTService,
) {
	log.Println("loaded AuthRoute")

	authRoute := router.Group("/auth")

	authRoute.Post("login", func(c fiber.Ctx) error {
		return loginHandler(c, jwtEngine)
	})

	authRoute.Get("self", func(c fiber.Ctx) error {
		return tokenCheckHandler(c, jwtEngine)
	})
}
