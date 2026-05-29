package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Headline struct {
	Title       string    `json:"title"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
}

// Query describes which Google News feed to fetch.
//
// Country selects the country edition (ISO 3166-1 alpha-2). Empty means US.
// Location narrows the feed to a city/state/region; empty means top headlines.
// N caps the number of items returned; 0 or negative means no cap.
type Query struct {
	Country  string
	Location string
	N        int
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

// FetchTopHeadlines is a convenience wrapper for the default US edition.
func FetchTopHeadlines(ctx context.Context, n int) ([]Headline, error) {
	return Fetch(ctx, Query{N: n})
}

// Fetch retrieves headlines for the given Query using the default HTTP client.
func Fetch(ctx context.Context, q Query) ([]Headline, error) {
	return fetchWith(ctx, http.DefaultClient, "https://news.google.com", q)
}

// fetchWith builds the feed URL for q against base and fetches it.
// base is the Google News origin (overridable for tests).
//
// When q.Location is set, the geo section endpoint is tried first; if it
// returns zero items, we fall back to the search endpoint so the user gets
// something rather than an empty list.
func fetchWith(ctx context.Context, client *http.Client, base string, q Query) ([]Headline, error) {
	ed, err := resolveEdition(q.Country)
	if err != nil {
		return nil, err
	}

	primary := buildFeedURL(base, ed, q.Location, false)
	items, err := fetchURL(ctx, client, primary)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 && q.Location != "" {
		fallback := buildFeedURL(base, ed, q.Location, true)
		items, err = fetchURL(ctx, client, fallback)
		if err != nil {
			return nil, err
		}
	}
	return limit(items, q.N), nil
}

// buildFeedURL constructs a Google News RSS URL.
//
//   - No location: /rss (top headlines for the edition)
//   - Location, useSearch=false: /rss/headlines/section/geo/<location>
//   - Location, useSearch=true:  /rss/search?q=<location>
//
// All forms append the edition's hl/gl/ceid params.
func buildFeedURL(base string, ed edition, location string, useSearch bool) string {
	params := url.Values{
		"hl":   {ed.hl},
		"gl":   {ed.gl},
		"ceid": {ed.ceid},
	}
	switch {
	case location == "":
		return base + "/rss?" + params.Encode()
	case useSearch:
		params.Set("q", location)
		return base + "/rss/search?" + params.Encode()
	default:
		return base + "/rss/headlines/section/geo/" + url.PathEscape(location) + "?" + params.Encode()
	}
}

func fetchURL(ctx context.Context, client *http.Client, url string) ([]Headline, error) {
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

	out := make([]Headline, 0, len(feed.Channel.Items))
	for _, it := range feed.Channel.Items {
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

func limit(items []Headline, n int) []Headline {
	if n > 0 && n < len(items) {
		return items[:n]
	}
	return items
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
