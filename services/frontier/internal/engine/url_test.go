package engine_test

import (
	"testing"

	"github.com/umutto/birdwatcher/services/frontier/internal/engine"
)

func TestNormalizeURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{
			name:   "root url without slash",
			rawURL: "http://example.com",
			want:   "http://example.com/",
		},
		{
			name:   "uppercase scheme and host",
			rawURL: "HTTP://EXAMPLE.COM/birds",
			want:   "http://example.com/birds",
		},
		{
			name:   "preserve case-sensitive path and query",
			rawURL: "https://example.com/wiki/Bald_Eagle?tag=BirdSightings",
			want:   "https://example.com/wiki/Bald_Eagle?tag=BirdSightings",
		},
		{
			name:   "subpath trailing slash removed",
			rawURL: "http://example.com/birds/",
			want:   "http://example.com/birds",
		},
		{
			name:   "strip default http port 80",
			rawURL: "http://example.com:80/birds",
			want:   "http://example.com/birds",
		},
		{
			name:   "preserve non-default port",
			rawURL: "http://localhost:8080/birds",
			want:   "http://localhost:8080/birds",
		},
		{
			name:   "strip url fragment",
			rawURL: "http://example.com/birds#sightings",
			want:   "http://example.com/birds",
		},
		{
			name:   "sort query parameters",
			rawURL: "http://example.com/search?b=2&a=1",
			want:   "http://example.com/search?a=1&b=2",
		},
		{
			name:    "unsupported scheme",
			rawURL:  "ftp://example.com/file",
			wantErr: true,
		},
		{
			name:    "empty host",
			rawURL:  "http:///path",
			wantErr: true,
		},
		{
			name:    "invalid url format",
			rawURL:  "://bad-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := engine.NormalizeURL(tt.rawURL)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NormalizeURL(%q) error = %v, wantErr %v", tt.rawURL, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("NormalizeURL(%q) = %q, want %q", tt.rawURL, got, tt.want)
			}
		})
	}
}
