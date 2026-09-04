package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazugmx/narubox-bot/internal/auth/jwt"
)

/*
- DO NOT access queries through routes. use on AuthHandler

/api/v1
- /auth
  * POST /register
  * POST /login
  * GET  /verify
- /users
  * GET  /me
*/

func Route(router fiber.Router, pool *pgxpool.Pool, jwtEngine *jwt.Service) {
	handler := NewAuthHandler(pool, jwtEngine)

	auth := router.Group("/auth")
	requireAuth := auth.Group("", jwtEngine.AuthMiddleware)

	auth.Post("/register", handler.registerHandler)

	auth.Post("/login", handler.loginHandler)

	// requireAuth.Get("/verify", handler.verifyHandler)

	requireAuth.Get("/users/me", handler.getSelfHandler)
}
