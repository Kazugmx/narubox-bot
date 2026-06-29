package Auth

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Kazugmx/narubox-bot/db"
	jwtOperator "github.com/Kazugmx/narubox-bot/internal/auth"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	jwtEngine *jwtOperator.JWTService
	query     *db.Queries
}

func NewAuthHandler(jwtEngine *jwtOperator.JWTService, query *db.Queries) *AuthHandler {
	return &AuthHandler{
		jwtEngine: jwtEngine,
		query:     query,
	}
}

func (authHandler *AuthHandler) loginHandler(c fiber.Ctx) error {
	req := c.Req().Body()
	var login_request UserLoginPayload
	err := json.Unmarshal(req, &login_request)
	if err != nil {
		slog.Error("error:", slog.Any("error", err))
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "Invalid Request payload.",
		})
	}

	challTarget, err := authHandler.query.GetAuthData(
		context.Background(),
		pgtype.Text{String: login_request.Username, Valid: true},
	)

	if err != nil {
		slog.Error("error:", slog.Any("error", err))
		return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
			"error": "Invalid username or password.",
		})
	}

	slog.Info("chall hash", slog.String("password", challTarget.Password))

	err = bcrypt.CompareHashAndPassword(
		[]byte(challTarget.Password),
		[]byte(login_request.Password),
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

	req_struct := c.Body()
	token := challToken{}
	json.Unmarshal(req_struct, &token)

	claims, err := authHandler.jwtEngine.VerifyJwtToken(token.Token)
	if err != nil {
		slog.Error("error:", slog.Any("error", err))
		return c.Status(fiber.ErrUnauthorized.Code).JSON(fiber.Map{
			"error": "Invalid token.",
		})
	}

	return c.JSON(fiber.Map{
		"status": "successfuly executed tokenCheckHandler",
		"claims": claims,
	})
}
