package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestFetchTopHeadlines(t *testing.T) {
	body, err := os.ReadFile("testdata/sample.rss")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.Write(body)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	items, err := fetchFrom(ctx, srv.Client(), srv.URL, 10)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	want := []Headline{
		{Title: "Senate passes landmark bill", Source: "Reuters", URL: "https://example.com/article-1"},
		{Title: "Storm sweeps East Coast", Source: "The Associated Press", URL: "https://example.com/article-2"},
		{Title: "Tech earnings surprise analysts", Source: "CNBC", URL: "https://example.com/article-3"},
	}
	for i, w := range want {
		got := items[i]
		if got.Title != w.Title {
			t.Errorf("item %d title: got %q want %q", i, got.Title, w.Title)
		}
		if got.Source != w.Source {
			t.Errorf("item %d source: got %q want %q", i, got.Source, w.Source)
		}
		if got.URL != w.URL {
			t.Errorf("item %d url: got %q want %q", i, got.URL, w.URL)
		}
		if got.PublishedAt.IsZero() {
			t.Errorf("item %d: published_at was not parsed", i)
		}
	}
}

func TestFetchTopHeadlines_LimitN(t *testing.T) {
	body, err := os.ReadFile("testdata/sample.rss")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(body)
	}))
	defer srv.Close()

	items, err := fetchFrom(context.Background(), srv.Client(), srv.URL, 2)
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items with n=2, got %d", len(items))
	}
}

func TestStripSourceSuffix(t *testing.T) {
	cases := []struct {
		title, source, want string
	}{
		{"Senate passes landmark bill - Reuters", "Reuters", "Senate passes landmark bill"},
		{"Tech earnings surprise analysts", "CNBC", "Tech earnings surprise analysts"},
		{"Headline with - hyphen - The Verge", "The Verge", "Headline with - hyphen"},
		{"No source available", "", "No source available"},
	}
	for _, c := range cases {
		got := stripSourceSuffix(c.title, c.source)
		if got != c.want {
			t.Errorf("stripSourceSuffix(%q, %q) = %q, want %q", c.title, c.source, got, c.want)
		}
	}
}
