package Auth

import (
	"log/slog"
	"strings"

	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/svc/auth/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

func NewAuthHandler(jwtEngine *jwtOperator.JWTService, query *db.Queries) *AuthHandler {
	return &AuthHandler{
		jwtEngine: jwtEngine,
		query:     query,
	}
}

func (authHandler *AuthHandler) loginHandler(c fiber.Ctx) error {
	var loginRequest UserLoginPayload
	if err := c.Bind().Body(&loginRequest); err != nil {
		slog.Warn("invalid login payload", slog.Any("error", err))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Request payload.",
		})
	}
	loginRequest.Username = strings.TrimSpace(loginRequest.Username)
	if loginRequest.Username == "" || loginRequest.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required.",
		})
	}

	challTarget, err := authHandler.query.GetAuthData(
		c.Context(),
		pgtype.Text{String: loginRequest.Username, Valid: true},
	)

	if err != nil {
		slog.Error("error:", slog.Any("error", err))
		return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
			"error": "Invalid username or password.",
		})
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(challTarget.Password),
		[]byte(loginRequest.Password),
	)
	if err != nil {
		slog.Error("error comparing passwords:", slog.Any("error", err))
		return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
			"error": "Invalid username or password.",
		})
	}

	token, err := authHandler.jwtEngine.GenerateJwtToken(challTarget.ID)
	if err != nil {
		slog.Error("error generating token:", slog.Any("error", err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token.",
		})
	}

	return c.JSON(fiber.Map{
		"token": token,
	})
}

func (authHandler *AuthHandler) tokenCheckHandler(c fiber.Ctx) error {
	type challToken struct {
		Token string `json:"token"`
	}

	token := challToken{}
	if err := c.Bind().Body(&token); err != nil || strings.TrimSpace(token.Token) == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid Request payload.",
		})
	}

	claims, err := authHandler.jwtEngine.VerifyJwtToken(token.Token)
	if err != nil {
		slog.Error("error:", slog.Any("error", err))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid token.",
		})
	}

	return c.JSON(fiber.Map{
		"status": "successfuly executed tokenCheckHandler",
		"claims": claims,
	})
}
