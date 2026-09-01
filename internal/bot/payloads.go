package bot

import db "github.com/kazugmx/narubox-bot/internal/db/sqlc"

type BotHandler struct {
	query *db.Queries
}
