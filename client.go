package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const feedURL = "https://news.google.com/rss?hl=en-US&gl=US&ceid=US:en"

type Headline struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	PubDate string `xml:"pubDate"`
	Source  string `xml:"source"`
}

func FetchTopHeadlines(ctx context.Context, n int) ([]Headline, error) {
	return fetchFrom(ctx, http.DefaultClient, feedURL, n)
}

func fetchFrom(ctx context.Context, client *http.Client, url string, n int) ([]Headline, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("feed returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	var feed rssFeed
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse rss: %w", err)
	}

	items := feed.Channel.Items
	if n > 0 && n < len(items) {
		items = items[:n]
	}

	out := make([]Headline, 0, len(items))
	for _, it := range items {
		pub, _ := parsePubDate(it.PubDate)
		out = append(out, Headline{
			Title:       stripSourceSuffix(it.Title, it.Source),
			Source:      it.Source,
			URL:         it.Link,
			PublishedAt: pub,
		})
	}
	return out, nil
}

func parsePubDate(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC1123, time.RFC1123Z, time.RFC822, time.RFC822Z} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized pubDate: %q", s)
}

// Google News appends " - <Source>" to titles. Strip it when present.
func stripSourceSuffix(title, source string) string {
	if source == "" {
		return title
	}
	suffix := " - " + source
	if strings.HasSuffix(title, suffix) {
		return strings.TrimSuffix(title, suffix)
	}
	return title
}
