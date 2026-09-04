package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
	"runtime"

	"github.com/alexedwards/argon2id"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kazugmx/narubox-bot/internal/auth/jwt"
	"github.com/kazugmx/narubox-bot/internal/auth/pass"
	argon2Sv "github.com/kazugmx/narubox-bot/internal/auth/pass/argon2"
	db "github.com/kazugmx/narubox-bot/internal/db/sqlc"
	"github.com/kazugmx/narubox-bot/internal/misc"
)

func JSONRequired(c fiber.Ctx, t any) error {
	if !c.Is("json") {
		return c.Status(fiber.StatusUnsupportedMediaType).JSON(fiber.Map{
			"error": "content-type must be application/json",
		})
	}

	if err := json.Unmarshal(c.Req().Body(), t); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	return nil
}

func NewAuthHandler(pool *pgxpool.Pool, jwtService *jwt.Service) *Handler {
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

	return &Handler{
		query:      queries,
		argon2sv:   argon2Sv.NewArgon2Service(params),
		dummyHash:  dummyHash,
		jwtService: jwtService,
	}
}

func (authHandler *Handler) loginHandler(c fiber.Ctx) error {
	ctx := c.Context()
	var request LoginRequestPayload
	if err := json.Unmarshal(c.Req().Body(), &request); err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "invalid login request",
		})
	}

	userRecord, err := authHandler.query.GetPassHashByUsername(ctx, request.Username)
	found := true

	// handle missing credentials
	if errors.Is(err, pgx.ErrNoRows) {
		found = false

		userRecord = db.GetPassHashByUsernameRow{
			ID:           uuid.Nil,
			PasswordHash: authHandler.dummyHash,
		}
	} else if err != nil {
		slog.Error("failed to query authentication data", "error", err)

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	// execute verification
	isSuccess := pass.Verify(
		request.Password,
		userRecord.PasswordHash,
		authHandler.argon2sv,
	)

	if !(isSuccess.OK && found) {
		slog.Info("failed for login challenge", "targetUser", request.Username)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "challenge failed",
		})
	}

	// update password
	if isSuccess.NewHash != nil {
		if err := authHandler.query.UpdatePassword(ctx, db.UpdatePasswordParams{
			ID:           userRecord.ID,
			PasswordHash: *isSuccess.NewHash,
		}); err != nil {
			slog.Error("failed to alter passHash with argon2id", "error", err)
		}
	}

	// return success w/jwt,cookie
	// TODO: implement jwt generation logic
	var token string
	token = "TODO_implementJWTLogic"

	if err := authHandler.query.UpdateLoginTime(ctx, userRecord.ID); err != nil {
		slog.Error("failed to update login time", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error",
		})
	}

	slog.Info("passed login challenge", "targetUser", request.Username)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "ok",
		"message": "login success",
		"token":   token,
	})
}

func (authHandler *Handler) registerHandler(c fiber.Ctx) error {
	var request RegistrationRequestPayload
	if err := JSONRequired(c, &request); err != nil {
		return err
	}

	if request.Username == "" ||
		request.MailAddress == "" ||
		request.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing_required_fields",
		})
	}

	slog.Info("debug message for register", "payload", request)

	if !pass.ValidateMailAddressRule(request.MailAddress) {
		slog.Info("debug message", "mail", pass.ValidateMailAddressRule(request.MailAddress))
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "mailaddress_format_violation",
		})
	}

	if !pass.ValidatePasswordRule(request.Password) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "password_policy_violation",
		})
	}

	hash, err := authHandler.argon2sv.GenerateHash(request.Password)
	if err != nil {
		slog.Error("failed to generate passHash with Argon2id", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error.",
		})
	}

	// upcoming data
	userData, err := authHandler.query.CreateUser(c.Context(), db.CreateUserParams{
		Username:     request.Username,
		Email:        request.MailAddress,
		PasswordHash: hash,
	})
	if err != nil {
		slog.Error("failed user registration", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "internal server error.",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":   "ok",
		"userID":   userData.ID.String(),
		"username": userData.Username,
	})

}

func (authHandler *Handler) verifyHandler(c fiber.Ctx) error {
	jwt.GetClaims(c)
	return misc.NotImplemented(c)
}

func (authHandler *Handler) getSelfHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}
