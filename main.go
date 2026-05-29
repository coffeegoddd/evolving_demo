package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	count := flag.Int("n", 10, "number of headlines to show")
	asJSON := flag.Bool("json", false, "emit results as JSON")
	showURL := flag.Bool("url", false, "include article URLs in text output")
	country := flag.String("country", "", "country edition (ISO 3166-1 alpha-2, e.g. US, GB, FR); default US")
	location := flag.String("location", "", "narrow headlines to a city, state, region, or country name (e.g. Boston, California)")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	q := Query{Country: *country, Location: *location, N: *count}
	if err := run(ctx, q, *asJSON, *showURL); err != nil {
		fmt.Fprintln(os.Stderr, "usnews:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, q Query, asJSON, showURL bool) error {
	items, err := Fetch(ctx, q)
	if err != nil {
		return err
	}
	if asJSON {
		return renderJSON(os.Stdout, items)
	}
	return renderText(os.Stdout, items, time.Now(), showURL)
}
