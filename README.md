# usnews

A tiny Go CLI that prints top news headlines from the public Google News RSS feed. No API key required. Default edition is US English; flags let you switch country editions or narrow to a place.

## Build

```bash
go build -o usnews .
```

## Usage

```bash
./usnews                              # top 10 US headlines (default)
./usnews -n 5                         # top 5
./usnews -json                        # JSON output
./usnews -url                         # include article URLs

./usnews -country GB                  # UK edition
./usnews -country FR -n 5             # French edition, top 5

./usnews -location "Boston"           # headlines focused on Boston (US edition)
./usnews -location "California"       # state-level focus
./usnews -country FR -location "Marseille"  # French edition, Marseille focus
```

### Supported country codes

ISO 3166-1 alpha-2: `US`, `GB`, `CA`, `AU`, `IE`, `IN`, `NZ`, `FR`, `DE`, `ES`, `IT`, `JP`, `BR`, `MX`. Unknown codes are rejected with the list. Case-insensitive.

### How `-location` works

The location string is passed through to Google News' geo section endpoint (`/rss/headlines/section/geo/<location>`), so Google handles disambiguation server-side — "Boston" resolves to Boston, MA by default in the US edition; switch editions to disambiguate elsewhere. If the geo section returns no items, the CLI falls back once to a search query so you get something rather than silence.

`-location` composes with `-country`: the country sets the edition's language and ranking, the location narrows the subject.

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

Fetches Google News RSS. Google News is the upstream aggregator; individual headlines link back to their original publishers.
