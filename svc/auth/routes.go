package Auth

import (
	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/svc/auth/jwt"
	"github.com/gofiber/fiber/v3"
)

func Route(
	router fiber.Router,
	query *db.Queries,
	jwtEngine *jwtOperator.JWTService,
) {
	authRoute := router.Group("/auth")
	handler := NewAuthHandler(jwtEngine, query)

	authRoute.Post("login", handler.loginHandler)
	authRoute.Get("self", handler.tokenCheckHandler)
}
