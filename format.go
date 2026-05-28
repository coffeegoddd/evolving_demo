package main

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

func renderText(w io.Writer, items []Headline, now time.Time, showURL bool) error {
	for i, h := range items {
		when := relativeTime(now, h.PublishedAt)
		if _, err := fmt.Fprintf(w, "%2d. %s — %s (%s)\n", i+1, h.Title, h.Source, when); err != nil {
			return err
		}
		if showURL && h.URL != "" {
			if _, err := fmt.Fprintf(w, "    %s\n", h.URL); err != nil {
				return err
			}
		}
	}
	return nil
}

func renderJSON(w io.Writer, items []Headline) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(items)
}

func relativeTime(now, then time.Time) string {
	if then.IsZero() {
		return "unknown"
	}
	d := now.Sub(then)
	if d < 0 {
		return "just now"
	}
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}
