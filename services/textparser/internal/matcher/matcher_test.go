package matcher_test

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/umutto/birdwatcher/services/textparser/internal/catalog"
	"github.com/umutto/birdwatcher/services/textparser/internal/matcher"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("empty species list", func(t *testing.T) {
		t.Parallel()

		_, err := matcher.New(nil)
		if err == nil {
			t.Fatal("New(nil) expected error, got nil")
		}
	})

	t.Run("valid species dictionary", func(t *testing.T) {
		t.Parallel()

		m, err := matcher.New(testSpecies())
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}

		if m == nil {
			t.Fatal("New() returned nil matcher")
		}
	})
}

func TestMatcher_Match(t *testing.T) {
	t.Parallel()

	m, err := matcher.New(testSpecies())
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		title string
		body  string
		want  []matcher.Sighting
	}{
		{
			name:  "empty inputs",
			title: "",
			body:  "   \n\t ",
			want:  nil,
		},
		{
			name:  "normalizes casing whitespace and punctuation across shared plates",
			title: "",
			body:  "Spotted a  BALD\n\tEAGLE. near the river",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 11,
					Confidence:         1.0,
					Snippet:            "spotted a bald eagle. near the river",
					MatchedText:        "bald eagle",
				},
				{
					SpeciesPlateNumber: 31,
					Confidence:         1.0,
					Snippet:            "spotted a bald eagle. near the river",
					MatchedText:        "bald eagle",
				},
			},
		},
		{
			name:  "enforces whole word boundaries",
			title: "",
			body:  "A red balloon floated over the saloon",
			want:  nil,
		},
		{
			name:  "prefers leftmost longest match",
			title: "",
			body:  "The Black Vulture soared overhead",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 106,
					Confidence:         1.0,
					Snippet:            "the black vulture soared overhead",
					MatchedText:        "black vulture",
				},
			},
		},
		{
			name:  "prioritizes title over body and deduplicates repeated mentions",
			title: "Bald Eagle Guide",
			body:  "The Haliaeetus leucocephalus and Common Loon live near lakes. Another Bald Eagle flew by.",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 11,
					Confidence:         1.0,
					Snippet:            "bald eagle guide",
					MatchedText:        "bald eagle",
				},
				{
					SpeciesPlateNumber: 31,
					Confidence:         1.0,
					Snippet:            "bald eagle guide",
					MatchedText:        "bald eagle",
				},
				{
					SpeciesPlateNumber: 306,
					Confidence:         1.0,
					Snippet:            "the haliaeetus leucocephalus and common loon live near lakes. another bald eagle flew by.",
					MatchedText:        "common loon",
				},
			},
		},
		{
			name:  "folds curly apostrophes",
			title: "",
			body:  "It\u2019s a Cooper\u2019s Hawk perched near a Chuck-will\u2019s-widow.",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 36,
					Confidence:         1.0,
					Snippet:            "it's a cooper's hawk perched near a chuck-will's-widow.",
					MatchedText:        "cooper's hawk",
				},
				{
					SpeciesPlateNumber: 52,
					Confidence:         1.0,
					Snippet:            "it's a cooper's hawk perched near a chuck-will's-widow.",
					MatchedText:        "chuck-will's-widow",
				},
			},
		},
		{
			name:  "matches dropped apostrophe and possessive",
			title: "Coopers Hawk sighting",
			body:  "We also saw a Harris' Hawk and a Lewis Woodpecker.",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 36,
					Confidence:         1.0,
					Snippet:            "coopers hawk sighting",
					MatchedText:        "coopers hawk",
				},
				{
					SpeciesPlateNumber: 392,
					Confidence:         1.0,
					Snippet:            "we also saw a harris' hawk and a lewis woodpecker.",
					MatchedText:        "harris' hawk",
				},
				{
					SpeciesPlateNumber: 416,
					Confidence:         1.0,
					Snippet:            "we also saw a harris' hawk and a lewis woodpecker.",
					MatchedText:        "lewis woodpecker",
				},
			},
		},
		{
			name:  "matches possessive spelling variants",
			title: "",
			body:  "Spotted a Cooper Hawk alongside a Lewis's Woodpecker.",
			want: []matcher.Sighting{
				{
					SpeciesPlateNumber: 36,
					Confidence:         1.0,
					Snippet:            "spotted a cooper hawk alongside a lewis's woodpecker.",
					MatchedText:        "cooper hawk",
				},
				{
					SpeciesPlateNumber: 416,
					Confidence:         1.0,
					Snippet:            "spotted a cooper hawk alongside a lewis's woodpecker.",
					MatchedText:        "lewis's woodpecker",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := m.Match(tt.title, tt.body)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Match() = %+v, want %+v", got, tt.want)
			}
		})
	}

	t.Run("slices utf8-safe snippet window", func(t *testing.T) {
		t.Parallel()

		body := strings.Repeat("é", 40) + " Common Loon " + strings.Repeat("—", 30)
		got := m.Match("", body)
		if len(got) != 1 || got[0].SpeciesPlateNumber != 306 {
			t.Fatalf("Match() = %+v, want single sighting for plate 306", got)
		}

		snippet := got[0].Snippet
		if !utf8.ValidString(snippet) {
			t.Errorf("Snippet is not valid UTF-8: %q", snippet)
		}
		if !strings.Contains(snippet, "common loon") {
			t.Errorf("Snippet %q does not contain matched species", snippet)
		}
		if len(snippet) >= len(body) {
			t.Errorf("Snippet length = %d, want truncated below %d", len(snippet), len(body))
		}
	})
}

func testSpecies() []catalog.Species {
	return []catalog.Species{
		{
			PlateNumber:    11,
			CommonName:     "Bald Eagle",
			ScientificName: "Haliaeetus leucocephalus",
			Aliases:        []string{"bald eagle", "bird of washington"},
		},
		{
			PlateNumber:    31,
			CommonName:     "Bald Eagle",
			ScientificName: "Haliaeetus leucocephalus",
		},
		{
			PlateNumber:    36,
			CommonName:     "Cooper's Hawk",
			ScientificName: "Astur cooperii",
		},
		{
			PlateNumber:    52,
			CommonName:     "Chuck-will's-widow",
			ScientificName: "Antrostomus carolinensis",
		},
		{
			PlateNumber:    106,
			CommonName:     "Black Vulture",
			ScientificName: "Coragyps atratus",
		},
		{
			PlateNumber:    306,
			CommonName:     "Common Loon",
			ScientificName: "Gavia immer",
			Aliases:        []string{"loon"},
		},
		{
			PlateNumber:    317,
			CommonName:     "Surf Scoter",
			ScientificName: "Melanitta perspicillata",
			Aliases:        []string{"black"},
		},
		{
			PlateNumber:    392,
			CommonName:     "Harris's Hawk",
			ScientificName: "Parabuteo unicinctus",
		},
		{
			PlateNumber:    416,
			CommonName:     "Hairy Woodpecker",
			ScientificName: "Leuconotopicus villosus",
			Aliases:        []string{"Lewis' Woodpecker"},
		},
	}
}
