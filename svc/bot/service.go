package bot

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Kazugmx/narubox-bot/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const hubURL = "https://pubsubhubbub.appspot.com/subscribe"

type Service struct {
	query   *db.Queries
	client  *http.Client
	apiKey  string
	origin  string
	logger  *slog.Logger
	mu      sync.Mutex
	process map[string]struct{}
}

func NewService(query *db.Queries, client *http.Client, apiKey, origin string, logger *slog.Logger) *Service {
	if client == nil {
		client = http.DefaultClient
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Service{query: query, client: client, apiKey: apiKey, origin: strings.TrimRight(origin, "/"), logger: logger, process: make(map[string]struct{})}
}

func (s *Service) ListBots(ctx context.Context, ownerID int32) ([]BotResponse, error) {
	bots, err := s.query.ListBots(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	result := make([]BotResponse, 0, len(bots))
	for _, bot := range bots {
		result = append(result, BotResponse{BotID: bot.ID.String(), Label: bot.Label, MentionRoleID: bot.MentionRoleID})
	}
	return result, nil
}

func (s *Service) RegisterBot(ctx context.Context, req RegisterRequest, ownerID int32) (RegisterResponse, error) {
	req.BotLabel = strings.TrimSpace(req.BotLabel)
	if req.BotLabel == "" || len(req.BotLabel) > 50 || req.MentionRoleID == "" {
		return RegisterResponse{}, errors.New("invalid bot registration")
	}
	webhook, err := parseHTTPURL(req.WebhookURL)
	if err != nil {
		return RegisterResponse{}, err
	}
	exists, err := s.query.BotLabelExists(ctx, ownerID, req.BotLabel)
	if err != nil {
		return RegisterResponse{}, err
	}
	if exists {
		return RegisterResponse{}, errors.New("bot label already exists")
	}

	botID := uuid.New()
	if err := s.query.CreateBot(ctx, db.Bot{ID: botID, OwnerID: ownerID, Label: req.BotLabel, WsURL: webhook.String(), MentionRoleID: req.MentionRoleID}); err != nil {
		return RegisterResponse{}, err
	}
	if err := s.sendWebhook(ctx, webhook.String(), map[string]string{"content": "Hello world, " + mentionText(req.MentionRoleID) + " !"}); err != nil {
		return RegisterResponse{}, fmt.Errorf("send bot test webhook: %w", err)
	}
	return RegisterResponse{Success: true, BotID: botID.String()}, nil
}

func (s *Service) UnregisterBot(ctx context.Context, botID string, ownerID int32) (bool, error) {
	id, err := db.ParseBotID(botID)
	if err != nil {
		return false, err
	}
	return s.query.DeleteBot(ctx, id, ownerID)
}

func (s *Service) GetBotInfo(ctx context.Context, botID string, ownerID int32) (BotInfoResponse, error) {
	id, err := db.ParseBotID(botID)
	if err != nil {
		return BotInfoResponse{}, err
	}
	bot, err := s.query.GetOwnedBot(ctx, id, ownerID)
	if err != nil {
		return BotInfoResponse{}, err
	}
	channels, err := s.query.ListBotChannels(ctx, id)
	if err != nil {
		return BotInfoResponse{}, err
	}
	return BotInfoResponse{BotInfo: BotResponse{BotID: bot.ID.String(), Label: bot.Label, MentionRoleID: bot.MentionRoleID}, Channels: channels}, nil
}

func (s *Service) RegisterChannel(ctx context.Context, botID string, req SubscribeRequest, ownerID int32) error {
	if req.ChannelID == "" || len(req.ChannelID) > 60 {
		return errors.New("invalid channel id")
	}
	id, err := db.ParseBotID(botID)
	if err != nil {
		return err
	}
	owned, err := s.query.IsBotOwner(ctx, id, ownerID)
	if err != nil || !owned {
		if err != nil {
			return err
		}
		return pgx.ErrNoRows
	}

	channel, err := s.query.GetChannel(ctx, req.ChannelID)
	if errors.Is(err, pgx.ErrNoRows) {
		channel = db.Channel{ChannelID: req.ChannelID, EndpointID: randomEndpointID()}
	} else if err != nil {
		return err
	}
	if req.Refresh || channel.EndpointID == "" {
		if channel.EndpointID == "" {
			channel.EndpointID = randomEndpointID()
		}
		if err := s.subscribe(ctx, channel); err != nil {
			return err
		}
		if err := s.query.UpsertChannel(ctx, channel); err != nil {
			return err
		}
	}
	return s.query.LinkChannel(ctx, id, req.ChannelID)
}

func (s *Service) DeleteChannel(ctx context.Context, botID, channelID string, ownerID int32) error {
	id, err := db.ParseBotID(botID)
	if err != nil {
		return err
	}
	owned, err := s.query.IsBotOwner(ctx, id, ownerID)
	if err != nil {
		return err
	}
	if !owned {
		return pgx.ErrNoRows
	}
	return s.query.DeleteBotChannel(ctx, id, channelID)
}

func (s *Service) RefreshChannels(ctx context.Context) error {
	channels, err := s.query.ListChannels(ctx)
	if err != nil {
		return err
	}
	for _, channel := range channels {
		if err := s.subscribe(ctx, channel); err != nil {
			return err
		}
		if err := s.query.UpsertChannel(ctx, channel); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) subscribe(ctx context.Context, channel db.Channel) error {
	callback := fmt.Sprintf("https://%s/api/v1/bot/pubsub/%s", s.origin, channel.EndpointID)
	topic := "https://www.youtube.com/xml/feeds/videos.xml?channel_id=" + url.QueryEscape(channel.ChannelID)
	values := url.Values{"hub.callback": {callback}, "hub.lease_seconds": {"864000"}, "hub.mode": {"subscribe"}, "hub.topic": {topic}, "hub.verify": {"sync"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, hubURL+"?"+values.Encode(), nil)
	if err != nil {
		return err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pubsub subscribe returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func (s *Service) NotifyToBots(ctx context.Context, videoID, endpointID string) {
	s.mu.Lock()
	if _, exists := s.process[videoID]; exists {
		s.mu.Unlock()
		return
	}
	s.process[videoID] = struct{}{}
	s.mu.Unlock()
	defer func() { s.mu.Lock(); delete(s.process, videoID); s.mu.Unlock() }()

	channel, err := s.findChannelByEndpoint(ctx, endpointID)
	if err != nil {
		s.logger.Error("invalid PubSub endpoint", "endpoint", endpointID, "error", err)
		return
	}
	previous, err := s.query.GetPreviousState(ctx, videoID)
	if errors.Is(err, pgx.ErrNoRows) {
		previous = string(StateError)
	} else if err != nil {
		return
	}
	data, err := s.youtubeVideoQuery(ctx, videoID)
	if err != nil {
		s.logger.Error("YouTube API request failed", "video", videoID, "error", err)
		return
	}
	for attempt := 1; data != nil && data.Type == previous && previous != string(StateVideo) && attempt <= 3; attempt++ {
		timer := time.NewTimer(time.Duration(attempt) * 20 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
		if next, queryErr := s.youtubeVideoQuery(ctx, videoID); queryErr == nil {
			data = next
		}
	}
	if data == nil {
		return
	}
	state, err := s.classifyStatus(ctx, *data)
	if err != nil || state == StateError {
		return
	}
	webhooks, err := s.query.ListWebhooks(ctx, channel)
	if err != nil {
		return
	}
	payload := buildWebhookPayload(*data, state, "")
	for _, webhook := range webhooks {
		payload.Content = buildMessage(*data, state, webhook.MentionRoleID)
		if err := s.sendWebhook(ctx, webhook.URL, payload); err != nil {
			s.logger.Error("failed to notify webhook", "url", webhook.URL, "error", err)
		}
	}
}

func (s *Service) findChannelByEndpoint(ctx context.Context, endpointID string) (string, error) {
	channels, err := s.query.ListChannels(ctx)
	if err != nil {
		return "", err
	}
	for _, channel := range channels {
		if channel.EndpointID == endpointID {
			return channel.ChannelID, nil
		}
	}
	return "", errors.New("endpoint not found")
}

func (s *Service) classifyStatus(ctx context.Context, data queryData) (DeliverState, error) {
	previous, err := s.query.GetPreviousState(ctx, data.VideoID)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := s.query.SetPreviousState(ctx, data.VideoID, data.Type); err != nil {
			return StateError, err
		}
		switch data.Type {
		case "upcoming":
			return StateUpcoming, nil
		case "live":
			return StateOnAir, nil
		case "none":
			return StateVideo, nil
		default:
			return StateError, nil
		}
	}
	if err != nil {
		return StateError, err
	}
	if previous == data.Type {
		return StateError, nil
	}
	switch previous {
	case "upcoming":
		if err := s.query.SetPreviousState(ctx, data.VideoID, "live"); err != nil {
			return StateError, err
		}
		return StateOnAir, nil
	case "live":
		if err := s.query.SetPreviousState(ctx, data.VideoID, "none"); err != nil {
			return StateError, err
		}
		return StateFinished, nil
	default:
		return StateError, nil
	}
}

func (s *Service) youtubeVideoQuery(ctx context.Context, videoID string) (*queryData, error) {
	values := url.Values{"part": {"snippet,liveStreamingDetails"}, "id": {videoID}, "key": {s.apiKey}}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://www.googleapis.com/youtube/v3/videos?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("YouTube API returned HTTP %d", resp.StatusCode)
	}
	var result videoListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Items) == 0 {
		return nil, nil
	}
	item := result.Items[0]
	published, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
	if err != nil {
		return nil, err
	}
	thumb := item.Snippet.Thumbnails.Default.URL
	for _, candidate := range []*thumbnail{item.Snippet.Thumbnails.Maxres, item.Snippet.Thumbnails.Standard, item.Snippet.Thumbnails.High, item.Snippet.Thumbnails.Medium} {
		if candidate != nil && candidate.URL != "" {
			thumb = candidate.URL
			break
		}
	}
	data := &queryData{Type: item.Snippet.LiveBroadcastContent, ChannelID: item.Snippet.ChannelID, VideoID: videoID, ChannelTitle: item.Snippet.ChannelTitle, Title: item.Snippet.Title, PublishedAt: published.Unix(), Thumbnail: thumb}
	if item.LiveStreamDetails != nil {
		data.StartTime = parseTime(item.LiveStreamDetails.ActualStartTime, item.LiveStreamDetails.ScheduledStartTime)
		data.EndTime = parseTime(item.LiveStreamDetails.ActualEndTime)
	}
	return data, nil
}

func parseTime(values ...string) *int64 {
	for _, value := range values {
		if value == "" {
			continue
		}
		if parsed, err := time.Parse(time.RFC3339, value); err == nil {
			unix := parsed.Unix()
			return &unix
		}
	}
	return nil
}

func buildWebhookPayload(data queryData, state DeliverState, message string) webhookPayload {
	videoURL := "https://www.youtube.com/watch?v=" + data.VideoID
	fields := make([]webhookField, 0, 3)
	addTime := func(name, format string, value *int64) {
		if value != nil {
			fields = append(fields, webhookField{Name: name, Value: fmt.Sprintf("<t:%d:%s>", *value, format)})
		}
	}
	switch state {
	case StateUpcoming:
		addTime("開始時刻(予定)", "F", data.StartTime)
	case StateOnAir:
		addTime("開始時刻", "F", data.StartTime)
	case StateFinished:
		addTime("開始日時", "F", data.StartTime)
		addTime("終了時刻", "T", data.EndTime)
	case StateVideo:
		addTime("投稿日時", "F", &data.PublishedAt)
	}
	fields = append(fields, webhookField{Name: "URL", Value: videoURL})
	return webhookPayload{Username: data.ChannelTitle + " - " + state.label(), Content: message, Embeds: []webhookEmbed{{Title: data.Title + " [" + state.label() + "]", URL: videoURL, Color: state.color(), Image: webhookImage{URL: data.Thumbnail}, Thumbnail: webhookImage{URL: data.Thumbnail}, Fields: fields}}}
}

func buildMessage(data queryData, state DeliverState, roleID string) string {
	message := mentionText(roleID)
	switch state {
	case StateUpcoming:
		return message + "配信枠が作成されました！"
	case StateOnAir:
		return message + "配信が開始されました！"
	case StateFinished:
		return "配信が終了しました。おつかれさまでした！"
	case StateVideo:
		return message + "動画が投稿されました！"
	default:
		return "エラーが発生しました！"
	}
}

func (s *Service) sendWebhook(ctx context.Context, target string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, target, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned HTTP %d", resp.StatusCode)
	}
	return nil
}

func mentionText(roleID string) string {
	if strings.HasPrefix(roleID, "ignore:") {
		return strings.TrimPrefix(roleID, "ignore:")
	}
	return "<@&" + roleID + "> "
}

func randomEndpointID() string {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return uuid.New().String()
	}
	return hex.EncodeToString(buf)
}

func parseHTTPURL(value string) (*url.URL, error) {
	parsed, err := url.Parse(value)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, errors.New("webhook URL must be an http or https URL")
	}
	return parsed, nil
}
