package bot

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	db "github.com/kazugmx/narubox-bot/internal/db/sqlc"
	"github.com/kazugmx/narubox-bot/internal/misc"
)

func NewBotHandler(pool *pgxpool.Pool) *BotHandler {
	queries := db.New(pool)
	return &BotHandler{
		query: queries,
	}
}

// feed Handling

func (botHandler *BotHandler) FeedVerifyHandler(c fiber.Ctx) error {
	//TODO: make channel claim verification
	endpointID := c.Params("endpoint_id")
	err := uuid.Validate(endpointID)
	parsedID, _ := uuid.Parse(endpointID)

	// insert logic here

	if err != nil {
		return c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"error": "invalid feed.",
		})
	}

	return c.SendString(parsedID.String())
}

func (botHandler *BotHandler) FeedDeliveryHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

//bot Handling

func (botHandler *BotHandler) GetBotWithIDHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (botHandler *BotHandler) CreateBotHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (botHandler *BotHandler) DeleteBotHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (botHandler *BotHandler) GetChannelListHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (botHandler *BotHandler) AddChannelHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}

func (botHandler *BotHandler) DeleteChannelHandler(c fiber.Ctx) error {
	return misc.NotImplemented(c)
}
