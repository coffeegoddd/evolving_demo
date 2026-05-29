package main

import (
	"fmt"
	"sort"
	"strings"
)

// edition holds the Google News URL params for a country edition.
// hl is the UI language code; gl is the country code; ceid is "<country>:<lang>".
type edition struct {
	hl   string
	gl   string
	ceid string
}

// editions maps ISO 3166-1 alpha-2 country codes to Google News editions.
// hl is not derivable from gl alone — multilingual countries (CA, CH, BE)
// pin a single language per edition, so we keep an explicit map.
var editions = map[string]edition{
	"US": {hl: "en-US", gl: "US", ceid: "US:en"},
	"GB": {hl: "en-GB", gl: "GB", ceid: "GB:en"},
	"CA": {hl: "en-CA", gl: "CA", ceid: "CA:en"},
	"AU": {hl: "en-AU", gl: "AU", ceid: "AU:en"},
	"IE": {hl: "en-IE", gl: "IE", ceid: "IE:en"},
	"IN": {hl: "en-IN", gl: "IN", ceid: "IN:en"},
	"NZ": {hl: "en-NZ", gl: "NZ", ceid: "NZ:en"},
	"FR": {hl: "fr", gl: "FR", ceid: "FR:fr"},
	"DE": {hl: "de", gl: "DE", ceid: "DE:de"},
	"ES": {hl: "es", gl: "ES", ceid: "ES:es"},
	"IT": {hl: "it", gl: "IT", ceid: "IT:it"},
	"JP": {hl: "ja", gl: "JP", ceid: "JP:ja"},
	"BR": {hl: "pt-BR", gl: "BR", ceid: "BR:pt-419"},
	"MX": {hl: "es-419", gl: "MX", ceid: "MX:es-419"},
}

// resolveEdition returns the edition for the given country code (case-insensitive).
// Empty code resolves to US. Unknown codes return an error listing supported codes.
func resolveEdition(country string) (edition, error) {
	if country == "" {
		return editions["US"], nil
	}
	code := strings.ToUpper(country)
	ed, ok := editions[code]
	if !ok {
		return edition{}, fmt.Errorf("unsupported country %q; supported: %s", country, supportedCountries())
	}
	return ed, nil
}

func supportedCountries() string {
	codes := make([]string, 0, len(editions))
	for c := range editions {
		codes = append(codes, c)
	}
	sort.Strings(codes)
	return strings.Join(codes, ", ")
}
