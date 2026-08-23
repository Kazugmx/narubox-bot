package bot

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"strings"

	jwtOperator "github.com/Kazugmx/narubox-bot/svc/auth/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
)

func Route(router fiber.Router, service *Service, jwtService *jwtOperator.JWTService) {
	botRoute := router.Group("/bot")
	botRoute.Get("/pubsub/:endpointID", pubsubChallenge)
	botRoute.Post("/pubsub/:endpointID", func(c fiber.Ctx) error {
		endpointID := c.Params("endpointID")
		var feed Feed
		if err := xml.Unmarshal(c.Body(), &feed); err != nil {
			return c.SendStatus(http.StatusOK)
		}
		if strings.Contains(string(c.Body()), "deleted-entry") {
			return c.SendStatus(http.StatusOK)
		}
		for _, entry := range feed.Entries {
			if entry.VideoID != "" {
				go service.NotifyToBots(c.Context(), entry.VideoID, endpointID)
			}
		}
		return c.SendStatus(http.StatusOK)
	})

	protected := botRoute.Group("", requireUser(jwtService))
	protected.Get("", func(c fiber.Ctx) error {
		bots, err := service.ListBots(c.Context(), userID(c))
		if err != nil {
			return serverError(c, err)
		}
		return c.JSON(bots)
	})
	protected.Get("/force", func(c fiber.Ctx) error {
		if err := service.RefreshChannels(c.Context()); err != nil {
			return serverError(c, err)
		}
		return c.SendStatus(http.StatusOK)
	})
	protected.Post("/register", func(c fiber.Ctx) error {
		var request RegisterRequest
		if err := c.Bind().Body(&request); err != nil {
			return badRequest(c, "Invalid request payload.")
		}
		result, err := service.RegisterBot(c.Context(), request, userID(c))
		if err != nil {
			return badRequest(c, err.Error())
		}
		return c.Status(http.StatusAccepted).JSON(result)
	})
	protected.Get("/:botID", func(c fiber.Ctx) error {
		result, err := service.GetBotInfo(c.Context(), c.Params("botID"), userID(c))
		if errors.Is(err, pgx.ErrNoRows) {
			return c.SendStatus(http.StatusNotFound)
		}
		if err != nil {
			return badRequest(c, err.Error())
		}
		return c.JSON(result)
	})
	protected.Delete("/:botID", func(c fiber.Ctx) error {
		deleted, err := service.UnregisterBot(c.Context(), c.Params("botID"), userID(c))
		if err != nil {
			return badRequest(c, err.Error())
		}
		return c.Status(http.StatusAccepted).JSON(fiber.Map{"success": deleted})
	})
	protected.Post("/:botID", func(c fiber.Ctx) error {
		var request SubscribeRequest
		if err := c.Bind().Body(&request); err != nil {
			return badRequest(c, "Invalid request payload.")
		}
		if err := service.RegisterChannel(c.Context(), c.Params("botID"), request, userID(c)); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return c.SendStatus(http.StatusForbidden)
			}
			return badRequest(c, err.Error())
		}
		return c.Status(http.StatusAccepted).JSON(fiber.Map{"success": true})
	})
	protected.Delete("/:botID/channel/:channelID", func(c fiber.Ctx) error {
		err := service.DeleteChannel(c.Context(), c.Params("botID"), c.Params("channelID"), userID(c))
		if errors.Is(err, pgx.ErrNoRows) {
			return c.SendStatus(http.StatusForbidden)
		}
		if err != nil {
			return badRequest(c, err.Error())
		}
		return c.Status(http.StatusAccepted).JSON(fiber.Map{"success": true})
	})
}

func pubsubChallenge(c fiber.Ctx) error {
	challenge := c.Query("hub.challenge")
	if challenge == "" {
		return c.SendStatus(http.StatusBadRequest)
	}
	return c.SendString(challenge)
}

type userContextKey struct{}

func requireUser(jwtService *jwtOperator.JWTService) fiber.Handler {
	return func(c fiber.Ctx) error {
		header := c.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			return c.SendStatus(http.StatusUnauthorized)
		}
		claims, err := jwtService.VerifyJwtToken(strings.TrimSpace(strings.TrimPrefix(header, "Bearer ")))
		if err != nil {
			return c.SendStatus(http.StatusUnauthorized)
		}
		id, err := strconv.ParseInt(claims.Subject, 10, 32)
		if err != nil {
			return c.SendStatus(http.StatusUnauthorized)
		}
		c.Locals(userContextKey{}, int32(id))
		return c.Next()
	}
}

func userID(c fiber.Ctx) int32 { return c.Locals(userContextKey{}).(int32) }

func badRequest(c fiber.Ctx, message string) error {
	return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": message})
}

func serverError(c fiber.Ctx, err error) error {
	return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
}
