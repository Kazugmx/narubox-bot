package Auth

import (
	"context"

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
	authRoute := router.Group("/auth")
	handler := NewAuthHandler(jwtEngine, query)

	authRoute.Post("login", handler.loginHandler)
	authRoute.Get("self", handler.tokenCheckHandler)
}
