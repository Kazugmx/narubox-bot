package bot

import (
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Route(
	router fiber.Router, pool *pgxpool.Pool,
) {
	handler := NewBotHandler(pool)

	feeds := router.Group("/feeds")

	feeds.Get("/:endpoint_id", handler.FeedVerifyHandler)
	feeds.Post("/:endpoint_id", handler.FeedDeliveryHandler)

	bots := router.Group("/bots")
	bots.Get("/:bot_id", handler.GetBotWithIDHandler)
	bots.Post("/", handler.CreateBotHandler)
	bots.Delete("/:bot_id", handler.DeleteBotHandler)

	bots.Get("/:bot_id/channels", handler.GetChannelListHandler)
	bots.Post("/:bot_id/channels/:channel_id", handler.AddChannelHandler)
	bots.Delete("/:bot_id/channels/:channel_id", handler.DeleteChannelHandler)
}
