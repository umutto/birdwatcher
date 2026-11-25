// Package matcher provides Aho-Corasick text scanning for detecting Audubon bird species.
package matcher

import (
	"cmp"
	"errors"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/umutto/birdwatcher/services/textparser/internal/catalog"
)

const (
	// exactMatchConfidence is a placeholder until future scoring or NLP is added;
	// for exact string matchers like Aho-Corasick it's always 1.0.
	exactMatchConfidence float32 = 1.0

	snippetPadding = 60
)

// apostropheReplacer folds common typographic apostrophes to ASCII.
var apostropheReplacer = strings.NewReplacer(
	"\u2019", "'",
	"\u2018", "'",
	"\u02bc", "'",
)

// Sighting represents a species match detected within a scanned page.
type Sighting struct {
	SpeciesPlateNumber int32
	Confidence         float32
	Snippet            string
	MatchedText        string
}

// Matcher scans page text against a compiled Aho-Corasick species automaton.
type Matcher struct {
	ac      ahocorasick.AhoCorasick
	targets map[string][]int32
}

// New builds a Matcher automaton from the provided catalog species dictionary.
func New(species []catalog.Species) (*Matcher, error) {
	// Estimated ~3 patterns per species (common name, scientific name, and ~1 alias).
	targets := make(map[string][]int32, len(species)*3)

	for _, sp := range species {
		addTarget(targets, sp.PlateNumber, sp.CommonName)
		addTarget(targets, sp.PlateNumber, sp.ScientificName)
		for _, alias := range sp.Aliases {
			addTarget(targets, sp.PlateNumber, alias)
		}
	}

	if len(targets) == 0 {
		return nil, errors.New("empty pattern dictionary")
	}

	builder := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		MatchOnlyWholeWords: true,
		MatchKind:           ahocorasick.LeftMostLongestMatch,
		DFA:                 true,
	})
	ac := builder.Build(slices.Collect(maps.Keys(targets)))

	return &Matcher{
		ac:      ac,
		targets: targets,
	}, nil
}

// Match scans title and bodyText and returns deduplicated species sightings.
func (m *Matcher) Match(title, bodyText string) []Sighting {
	seen := make(map[int32]Sighting)
	m.scan(title, seen)
	m.scan(bodyText, seen)

	sightings := slices.Collect(maps.Values(seen))
	slices.SortFunc(sightings, func(a, b Sighting) int {
		return cmp.Compare(a.SpeciesPlateNumber, b.SpeciesPlateNumber)
	})

	return sightings
}

// scan normalizes raw and records the first sighting for each matched species plate in seen.
func (m *Matcher) scan(raw string, seen map[int32]Sighting) {
	normalized := normalize(raw)
	if normalized == "" {
		return
	}

	for _, match := range m.ac.FindAll(normalized) {
		matchedText := normalized[match.Start():match.End()]
		plates := m.targets[matchedText]
		snippet := extractSnippet(normalized, match.Start(), match.End())

		for _, plate := range plates {
			if _, exists := seen[plate]; exists {
				continue
			}

			seen[plate] = Sighting{
				SpeciesPlateNumber: plate,
				Confidence:         exactMatchConfidence,
				Snippet:            snippet,
				MatchedText:        matchedText,
			}
		}
	}
}

// extractSnippet returns a UTF-8-safe context window around text[start:end].
func extractSnippet(text string, start, end int) string {
	left := max(0, start-snippetPadding)
	for left > 0 && !utf8.RuneStart(text[left]) {
		left--
	}

	right := min(len(text), end+snippetPadding)
	for right < len(text) && !utf8.RuneStart(text[right]) {
		right++
	}

	return strings.TrimSpace(text[left:right])
}

// addTarget normalizes raw and registers plate under each generated pattern variant if non-empty.
func addTarget(targets map[string][]int32, plate int32, raw string) {
	pattern := normalize(raw)
	if pattern == "" {
		return
	}

	for _, variant := range patternVariants(pattern) {
		if slices.Contains(targets[variant], plate) {
			continue
		}
		targets[variant] = append(targets[variant], plate)
	}
}

// normalize lowercases s, folds typographic apostrophes to ASCII, and collapses whitespace into single spaces.
func normalize(s string) string {
	s = apostropheReplacer.Replace(strings.ToLower(s))
	return strings.Join(strings.Fields(s), " ")
}

// patternVariants returns pattern along with natural possessive spelling variants.
func patternVariants(pattern string) []string {
	tokens := strings.Fields(pattern)
	variants := []string{pattern}

	for i, token := range tokens {
		for _, alt := range possessiveAlternatives(token) {
			variantTokens := slices.Clone(tokens)
			variantTokens[i] = alt
			variants = append(variants, strings.Join(variantTokens, " "))
		}
	}

	return variants
}

// possessiveAlternatives returns alternate possessive spellings for a single normalized token.
func possessiveAlternatives(token string) []string {
	if stem, ok := strings.CutSuffix(token, "'s"); ok && stem != "" {
		if strings.HasSuffix(stem, "s") {
			return []string{stem + "'", stem}
		}
		return []string{stem, stem + "s"}
	}
	if stem, ok := strings.CutSuffix(token, "'"); ok && stem != "" {
		return []string{stem + "'s", stem}
	}
	return nil
}
