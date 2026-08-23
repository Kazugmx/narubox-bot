package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Bot struct {
	ID            uuid.UUID
	OwnerID       int32
	Label         string
	WsURL         string
	MentionRoleID string
}

type Channel struct {
	ChannelID  string
	EndpointID string
}

type Webhook struct {
	URL           string
	MentionRoleID string
}

func (q *Queries) ListBots(ctx context.Context, ownerID int32) ([]Bot, error) {
	rows, err := q.db.Query(ctx, `
		SELECT id, owner_id, label, ws_url, mention_role_id
		FROM bot_table WHERE owner_id = $1 ORDER BY label`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bots []Bot
	for rows.Next() {
		var bot Bot
		if err := rows.Scan(&bot.ID, &bot.OwnerID, &bot.Label, &bot.WsURL, &bot.MentionRoleID); err != nil {
			return nil, err
		}
		bots = append(bots, bot)
	}
	return bots, rows.Err()
}

func (q *Queries) BotLabelExists(ctx context.Context, ownerID int32, label string) (bool, error) {
	var exists bool
	err := q.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM bot_table WHERE owner_id = $1 AND label = $2)`, ownerID, label).Scan(&exists)
	return exists, err
}

func (q *Queries) CreateBot(ctx context.Context, bot Bot) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO bot_table (id, owner_id, label, ws_url, mention_role_id)
		VALUES ($1, $2, $3, $4, $5)`,
		bot.ID, bot.OwnerID, bot.Label, bot.WsURL, bot.MentionRoleID)
	return err
}

func (q *Queries) GetOwnedBot(ctx context.Context, botID uuid.UUID, ownerID int32) (Bot, error) {
	var bot Bot
	err := q.db.QueryRow(ctx, `
		SELECT id, owner_id, label, ws_url, mention_role_id
		FROM bot_table WHERE id = $1 AND owner_id = $2`, botID, ownerID).
		Scan(&bot.ID, &bot.OwnerID, &bot.Label, &bot.WsURL, &bot.MentionRoleID)
	return bot, err
}

func (q *Queries) DeleteBot(ctx context.Context, botID uuid.UUID, ownerID int32) (bool, error) {
	result, err := q.db.Exec(ctx, `DELETE FROM bot_table WHERE id = $1 AND owner_id = $2`, botID, ownerID)
	return result.RowsAffected() > 0, err
}

func (q *Queries) ListBotChannels(ctx context.Context, botID uuid.UUID) ([]string, error) {
	rows, err := q.db.Query(ctx, `
		SELECT channel_id FROM channel_bot_tags WHERE bot_id = $1 ORDER BY channel_id`, botID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []string
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, rows.Err()
}

func (q *Queries) IsBotOwner(ctx context.Context, botID uuid.UUID, ownerID int32) (bool, error) {
	var exists bool
	err := q.db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM bot_table WHERE id = $1 AND owner_id = $2)`, botID, ownerID).Scan(&exists)
	return exists, err
}

func (q *Queries) GetChannel(ctx context.Context, channelID string) (Channel, error) {
	var channel Channel
	err := q.db.QueryRow(ctx, `
		SELECT channel_id, endpoint_id FROM reg_channel WHERE channel_id = $1`, channelID).
		Scan(&channel.ChannelID, &channel.EndpointID)
	return channel, err
}

func (q *Queries) UpsertChannel(ctx context.Context, channel Channel) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO reg_channel (channel_id, endpoint_id, last_update)
		VALUES ($1, $2, NOW())
		ON CONFLICT (channel_id) DO UPDATE SET endpoint_id = EXCLUDED.endpoint_id, last_update = NOW()`,
		channel.ChannelID, channel.EndpointID)
	return err
}

func (q *Queries) ListChannels(ctx context.Context) ([]Channel, error) {
	rows, err := q.db.Query(ctx, `SELECT channel_id, endpoint_id FROM reg_channel ORDER BY channel_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []Channel
	for rows.Next() {
		var channel Channel
		if err := rows.Scan(&channel.ChannelID, &channel.EndpointID); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, rows.Err()
}

func (q *Queries) LinkChannel(ctx context.Context, botID uuid.UUID, channelID string) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO channel_bot_tags (bot_id, channel_id) VALUES ($1, $2)
		ON CONFLICT (bot_id, channel_id) DO NOTHING`, botID, channelID)
	return err
}

func (q *Queries) DeleteBotChannel(ctx context.Context, botID uuid.UUID, channelID string) error {
	_, err := q.db.Exec(ctx, `DELETE FROM channel_bot_tags WHERE bot_id = $1 AND channel_id = $2`, botID, channelID)
	return err
}

func (q *Queries) GetPreviousState(ctx context.Context, videoID string) (string, error) {
	var state string
	err := q.db.QueryRow(ctx, `SELECT previous_state FROM on_air WHERE video_id = $1`, videoID).Scan(&state)
	return state, err
}

func (q *Queries) SetPreviousState(ctx context.Context, videoID, state string) error {
	_, err := q.db.Exec(ctx, `
		INSERT INTO on_air (video_id, previous_state) VALUES ($1, $2)
		ON CONFLICT (video_id) DO UPDATE SET previous_state = EXCLUDED.previous_state`, videoID, state)
	return err
}

func (q *Queries) ListWebhooks(ctx context.Context, channelID string) ([]Webhook, error) {
	rows, err := q.db.Query(ctx, `
		SELECT b.ws_url, b.mention_role_id
		FROM channel_bot_tags t JOIN bot_table b ON b.id = t.bot_id
		WHERE t.channel_id = $1`, channelID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var webhooks []Webhook
	for rows.Next() {
		var webhook Webhook
		if err := rows.Scan(&webhook.URL, &webhook.MentionRoleID); err != nil {
			return nil, err
		}
		webhooks = append(webhooks, webhook)
	}
	return webhooks, rows.Err()
}

func ParseBotID(value string) (uuid.UUID, error) {
	botID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid bot id: %w", err)
	}
	return botID, nil
}
