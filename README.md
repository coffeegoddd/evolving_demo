# usnews

A tiny Go CLI that prints the top US news headlines from the public Google News RSS feed. No API key required.

## Build

```bash
go build -o usnews .
```

## Usage

```bash
./usnews              # top 10 headlines (default)
./usnews -n 5         # top 5 headlines
./usnews -json        # emit results as JSON
./usnews -url         # include article URLs under each headline
```

## Example

```
 1. Senate passes landmark bill — Reuters (2h ago)
 2. Storm sweeps East Coast — The Associated Press (3h ago)
 ...
```

## Test

```bash
go test ./...
```

## Source

Fetches `https://news.google.com/rss?hl=en-US&gl=US&ceid=US:en`. Google News is the upstream aggregator; individual headlines link back to their original publishers.
