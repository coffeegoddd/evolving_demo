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
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := run(ctx, *count, *asJSON, *showURL); err != nil {
		fmt.Fprintln(os.Stderr, "usnews:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, count int, asJSON, showURL bool) error {
	items, err := FetchTopHeadlines(ctx, count)
	if err != nil {
		return err
	}
	if asJSON {
		return renderJSON(os.Stdout, items)
	}
	return renderText(os.Stdout, items, time.Now(), showURL)
}
