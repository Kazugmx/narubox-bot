package bot

import (
	"encoding/xml"
	"testing"
)

func TestFeedXML(t *testing.T) {
	var feed Feed
	err := xml.Unmarshal([]byte(`<feed xmlns="http://www.w3.org/2005/Atom" xmlns:yt="http://www.youtube.com/xml/schemas/2015"><entry><yt:videoId>video-1</yt:videoId><yt:channelId>channel-1</yt:channelId></entry></feed>`), &feed)
	if err != nil {
		t.Fatalf("xml.Unmarshal() error = %v", err)
	}
	if len(feed.Entries) != 1 || feed.Entries[0].VideoID != "video-1" || feed.Entries[0].ChannelID != "channel-1" {
		t.Fatalf("unexpected feed: %+v", feed)
	}
}

func TestBuildWebhookPayload(t *testing.T) {
	start := int64(1700000000)
	data := queryData{
		Type: "live", ChannelID: "channel-1", VideoID: "video-1",
		ChannelTitle: "Channel", Title: "Title", PublishedAt: 1699990000,
		StartTime: &start, Thumbnail: "https://example.test/thumb.jpg",
	}

	payload := buildWebhookPayload(data, StateOnAir, buildMessage(data, StateOnAir, "123"))
	if payload.Content != "<@&123> 配信が開始されました！" {
		t.Fatalf("payload.Content = %q", payload.Content)
	}
	if payload.Embeds[0].Color != 0xFF3838 || payload.Embeds[0].URL != "https://www.youtube.com/watch?v=video-1" {
		t.Fatalf("unexpected embed: %+v", payload.Embeds[0])
	}
}

func TestParseHTTPURL(t *testing.T) {
	if _, err := parseHTTPURL("ftp://example.test/webhook"); err == nil {
		t.Fatal("parseHTTPURL() error = nil for unsupported scheme")
	}
	if _, err := parseHTTPURL("https://discord.test/webhook"); err != nil {
		t.Fatalf("parseHTTPURL() error = %v", err)
	}
}
