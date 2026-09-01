package auth

import (
	"encoding/json"
	"log/slog"
	"runtime"

	"github.com/alexedwards/argon2id"
	"github.com/gofiber/fiber/v3"
	googleuuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	argon2Sv "github.com/kazugmx/narubox-bot/internal/auth/pass/argon2"
	db "github.com/kazugmx/narubox-bot/internal/db/sqlc"
	"github.com/kazugmx/narubox-bot/internal/misc"
)

func NewAuthHandler(pool *pgxpool.Pool) *AuthHandler {
	queries := db.New(pool)
	params := &argon2id.Params{
		Memory:      64 * 1024,
		Iterations:  2,
		Parallelism: uint8(runtime.NumCPU()),
		SaltLength:  16,
		KeyLength:   32,
	}

	dummyHash, _ := argon2id.CreateHash(
		"01a05f1d-97c7-7586-8aaf-12877db2df4b",
		params,
	)

	return &AuthHandler{
		query:     queries,
		argon2sv:  argon2Sv.NewArgon2Service(params),
		dummyHash: dummyHash,
	}
}

func (authHandler *AuthHandler) loginHandler(c fiber.Ctx) error {
	c.AcceptsJSON()

	ctx := c.Context()
	var request LoginRequestPayload
	if err := json.Unmarshal(c.Req().Body(), &request); err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "invalid login request",
		})
	}

	userRecord, err := authHandler.query.GetPassHashByUsername(ctx, request.Username)

	if err != nil {
		slog.Info("Failed login", "usertarget", request.Username)
		dummyID := googleuuid.MustParse("01a05f1d-97c7-7586-8aaf-12877db2df4b")

		userRecord = db.GetPassHashByUsernameRow{
			ID:           dummyID,
			PasswordHash: authHandler.dummyHash,
		}
	} else {
	}

	userID := userRecord.ID.String()

	_ = userID
	return c.Status(fiber.ErrForbidden.Code).JSON(fiber.Map{
		"error": "challenge failed",
	})
}

func (authHandler *AuthHandler) registerHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (authHandler *AuthHandler) verifyHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (authHandler *AuthHandler) getSelfHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}
