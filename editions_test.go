package main

import (
	"strings"
	"testing"
)

func TestResolveEdition(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantGL  string
		wantErr bool
	}{
		{"empty defaults to US", "", "US", false},
		{"uppercase", "GB", "GB", false},
		{"lowercase normalized", "fr", "FR", false},
		{"mixed case normalized", "Br", "BR", false},
		{"unknown errors", "ZZ", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ed, err := resolveEdition(c.input)
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", c.input)
				}
				if !strings.Contains(err.Error(), "supported:") {
					t.Errorf("error should list supported codes, got %q", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ed.gl != c.wantGL {
				t.Errorf("gl: got %q want %q", ed.gl, c.wantGL)
			}
		})
	}
}

func TestBuildFeedURL(t *testing.T) {
	us := editions["US"]
	fr := editions["FR"]

	cases := []struct {
		name      string
		ed        edition
		location  string
		useSearch bool
		wantParts []string // substrings the URL must contain
	}{
		{
			name:      "default US edition, no location",
			ed:        us,
			wantParts: []string{"https://news.google.com/rss?", "hl=en-US", "gl=US", "ceid=US%3Aen"},
		},
		{
			name:      "FR edition, no location",
			ed:        fr,
			wantParts: []string{"https://news.google.com/rss?", "hl=fr", "gl=FR", "ceid=FR%3Afr"},
		},
		{
			name:      "US edition with city location uses geo section",
			ed:        us,
			location:  "Boston",
			wantParts: []string{"/rss/headlines/section/geo/Boston?", "hl=en-US", "gl=US"},
		},
		{
			name:      "location with spaces is path-escaped",
			ed:        us,
			location:  "New York",
			wantParts: []string{"/rss/headlines/section/geo/New%20York?"},
		},
		{
			name:      "FR edition with city location",
			ed:        fr,
			location:  "Marseille",
			wantParts: []string{"/rss/headlines/section/geo/Marseille?", "hl=fr", "gl=FR"},
		},
		{
			name:      "search fallback URL",
			ed:        us,
			location:  "Springfield",
			useSearch: true,
			wantParts: []string{"/rss/search?", "q=Springfield", "hl=en-US"},
		},
		{
			name:      "search with spaces query-escapes",
			ed:        us,
			location:  "New York",
			useSearch: true,
			wantParts: []string{"q=New+York"},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := buildFeedURL("https://news.google.com", c.ed, c.location, c.useSearch)
			for _, want := range c.wantParts {
				if !strings.Contains(got, want) {
					t.Errorf("URL %q missing %q", got, want)
				}
			}
		})
	}
}
