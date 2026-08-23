package bot

type RegisterRequest struct {
	BotLabel      string `json:"botLabel"`
	WebhookURL    string `json:"wsUrl"`
	MentionRoleID string `json:"mentionRoleID"`
}

type RegisterResponse struct {
	Success bool   `json:"success"`
	BotID   string `json:"botID,omitempty"`
}

type BotResponse struct {
	BotID         string `json:"botID"`
	Label         string `json:"label"`
	MentionRoleID string `json:"mentionRoleID"`
}

type BotInfoResponse struct {
	BotInfo  BotResponse `json:"botInfo"`
	Channels []string    `json:"channels"`
}

type SubscribeRequest struct {
	ChannelID string `json:"channelID"`
	Refresh   bool   `json:"refresh"`
}

type Feed struct {
	Entries []FeedEntry `xml:"entry"`
}

type FeedEntry struct {
	VideoID   string `xml:"videoId"`
	ChannelID string `xml:"channelId"`
}

type videoListResponse struct {
	Items []videoItem `json:"items"`
}

type videoItem struct {
	Snippet           videoSnippet       `json:"snippet"`
	LiveStreamDetails *liveStreamDetails `json:"liveStreamingDetails"`
}

type videoSnippet struct {
	ChannelID            string     `json:"channelId"`
	Title                string     `json:"title"`
	ChannelTitle         string     `json:"channelTitle"`
	LiveBroadcastContent string     `json:"liveBroadcastContent"`
	PublishedAt          string     `json:"publishedAt"`
	Thumbnails           thumbnails `json:"thumbnails"`
}

type thumbnails struct {
	Default  thumbnail  `json:"default"`
	Medium   *thumbnail `json:"medium"`
	High     *thumbnail `json:"high"`
	Standard *thumbnail `json:"standard"`
	Maxres   *thumbnail `json:"maxres"`
}

type thumbnail struct {
	URL string `json:"url"`
}

type liveStreamDetails struct {
	ActualStartTime    string `json:"actualStartTime"`
	ScheduledStartTime string `json:"scheduledStartTime"`
	ActualEndTime      string `json:"actualEndTime"`
}

type queryData struct {
	Type         string
	ChannelID    string
	VideoID      string
	ChannelTitle string
	Title        string
	PublishedAt  int64
	StartTime    *int64
	EndTime      *int64
	Thumbnail    string
}

type DeliverState string

const (
	StateUpcoming DeliverState = "upcoming"
	StateOnAir    DeliverState = "live"
	StateFinished DeliverState = "finished"
	StateVideo    DeliverState = "video"
	StateError    DeliverState = "invalid"
)

func (s DeliverState) label() string {
	switch s {
	case StateUpcoming:
		return "配信予告"
	case StateOnAir:
		return "配信開始"
	case StateFinished:
		return "配信終了"
	case StateVideo:
		return "動画投稿"
	default:
		return "エラー"
	}
}

func (s DeliverState) color() int {
	switch s {
	case StateUpcoming:
		return 0x050E3C
	case StateOnAir:
		return 0xFF3838
	case StateFinished:
		return 0x1B3C53

	case StateVideo:
		return 0x6DC3BB
	default:
		return 0xFFFFFF
	}
}

type webhookPayload struct {
	Username string         `json:"username"`
	Embeds   []webhookEmbed `json:"embeds"`
	Content  string         `json:"content,omitempty"`
}

type webhookEmbed struct {
	Title     string         `json:"title"`
	URL       string         `json:"url"`
	Color     int            `json:"color"`
	Image     webhookImage   `json:"image"`
	Fields    []webhookField `json:"fields"`
	Thumbnail webhookImage   `json:"thumbnail"`
}

type webhookField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type webhookImage struct {
	URL string `json:"url"`
}
