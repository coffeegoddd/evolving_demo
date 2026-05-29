package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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

	items, err := fetchWith(ctx, srv.Client(), srv.URL, Query{N: 10})
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

	items, err := fetchWith(context.Background(), srv.Client(), srv.URL, Query{N: 2})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items with n=2, got %d", len(items))
	}
}

func TestFetchWith_CountryEdition(t *testing.T) {
	body, err := os.ReadFile("testdata/sample_uk.rss")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		w.Write(body)
	}))
	defer srv.Close()

	items, err := fetchWith(context.Background(), srv.Client(), srv.URL, Query{Country: "GB"})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if !strings.Contains(gotURL, "gl=GB") || !strings.Contains(gotURL, "hl=en-GB") {
		t.Errorf("expected GB edition params, got %q", gotURL)
	}
	if len(items) != 2 || items[0].Source != "BBC News" {
		t.Errorf("did not parse UK fixture: %+v", items)
	}
}

func TestFetchWith_LocationHitsGeoSection(t *testing.T) {
	body, err := os.ReadFile("testdata/sample_geo.rss")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Write(body)
	}))
	defer srv.Close()

	items, err := fetchWith(context.Background(), srv.Client(), srv.URL, Query{Location: "Boston"})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if gotPath != "/rss/headlines/section/geo/Boston" {
		t.Errorf("expected geo section path, got %q", gotPath)
	}
	if len(items) != 2 || items[0].Source != "The Boston Globe" {
		t.Errorf("did not parse geo fixture: %+v", items)
	}
}

func TestFetchWith_GeoEmptyFallsBackToSearch(t *testing.T) {
	emptyBody, err := os.ReadFile("testdata/sample_empty.rss")
	if err != nil {
		t.Fatalf("read empty fixture: %v", err)
	}
	geoBody, err := os.ReadFile("testdata/sample_geo.rss")
	if err != nil {
		t.Fatalf("read geo fixture: %v", err)
	}

	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if strings.HasPrefix(r.URL.Path, "/rss/headlines/section/geo/") {
			w.Write(emptyBody)
			return
		}
		w.Write(geoBody)
	}))
	defer srv.Close()

	items, err := fetchWith(context.Background(), srv.Client(), srv.URL, Query{Location: "Nowhereville"})
	if err != nil {
		t.Fatalf("fetch: %v", err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 requests (geo + search fallback), got %d: %v", len(paths), paths)
	}
	if paths[0] != "/rss/headlines/section/geo/Nowhereville" {
		t.Errorf("first request should be geo section, got %q", paths[0])
	}
	if paths[1] != "/rss/search" {
		t.Errorf("second request should be search, got %q", paths[1])
	}
	if len(items) == 0 {
		t.Error("expected fallback to surface search results, got 0 items")
	}
}

func TestFetchWith_UnknownCountryErrors(t *testing.T) {
	_, err := fetchWith(context.Background(), http.DefaultClient, "https://news.google.com", Query{Country: "XX"})
	if err == nil {
		t.Fatal("expected error for unknown country, got nil")
	}
	if !strings.Contains(err.Error(), "unsupported country") {
		t.Errorf("error message should mention unsupported country, got %q", err)
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
