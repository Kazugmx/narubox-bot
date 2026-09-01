package auth

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

/*
- DO NOT access queries through AuthHandler

/api/v1
- /auth
  * POST /register
  * POST /login
  * GET  /verify
- /users
  * GET  /me
*/

func Route(router fiber.Router, pool *pgxpool.Pool) {
	handler := NewAuthHandler(pool)

	auth := router.Group("/auth")

	auth.Post("/register", handler.registerHandler)

	auth.Post("/login", handler.loginHandler)
	auth.Get("/verify", handler.verifyHandler)

	router.Get("/users/me", handler.getSelfHandler)
}
