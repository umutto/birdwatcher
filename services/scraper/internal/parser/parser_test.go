package parser_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/umutto/birdwatcher/services/scraper/internal/parser"
)

const sampleHTML = `<!DOCTYPE html>
<html>
<head>
	<title>   Bald Eagle - Wildlife   </title>
	<script>console.log("script content should be skipped");</script>
	<style>body { color: red; }</style>
	<noscript><p>enable javascript</p></noscript>
</head>
<body>
	<h1>Bald Eagle</h1>
	<p>The <em>Bald Eagle</em> is a bird of prey found in North America.</p>
	<a href="/species/bald-eagle">Eagle Guide</a>
	<a href="https://example.com/gallery">Photo Gallery</a>
	<a href="/species/bald-eagle#habitat">Duplicate Link Stripping Fragment</a>
	<a href="mailto:info@example.com">Email Us</a>
	<a href="javascript:void(0)">Noop</a>
	<a href="#top">Back to top</a>
	<a>Anchor with no href</a>
	<img src="../images/bald-eagle.jpg" alt="Bald Eagle portrait">
	<img src="https://example.com/images/bald-eagle.jpg#view">
	<img src="data:image/png;base64,iVBORw0KGgo=">
</body>
</html>`

func TestParse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		baseURL string
		html    string
		want    *parser.Result
		wantErr bool
	}{
		{
			name:    "invalid base url scheme",
			baseURL: "ftp://example.com",
			html:    "<html><body>hello</body></html>",
			wantErr: true,
		},
		{
			name:    "invalid base url format",
			baseURL: "://bad-url",
			html:    "<html><body>hello</body></html>",
			wantErr: true,
		},
		{
			name:    "empty base url host",
			baseURL: "http:///path",
			html:    "<html><body>hello</body></html>",
			wantErr: true,
		},
		{
			name:    "empty html document",
			baseURL: "https://example.com",
			html:    "",
			want: &parser.Result{
				Title:     "",
				BodyText:  "",
				Links:     nil,
				ImageURLs: nil,
			},
		},
		{
			name:    "extracts title body links and images with normalization",
			baseURL: "https://example.com/articles/",
			html:    sampleHTML,
			want: &parser.Result{
				Title:    "Bald Eagle - Wildlife",
				BodyText: "Bald Eagle The Bald Eagle is a bird of prey found in North America. Eagle Guide Photo Gallery Duplicate Link Stripping Fragment Email Us Noop Back to top Anchor with no href",
				Links: []string{
					"https://example.com/species/bald-eagle",
					"https://example.com/gallery",
				},
				ImageURLs: []string{
					"https://example.com/images/bald-eagle.jpg",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parser.Parse(tt.baseURL, strings.NewReader(tt.html))
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if got.Title != tt.want.Title {
				t.Errorf("Title = %q, want %q", got.Title, tt.want.Title)
			}
			if got.BodyText != tt.want.BodyText {
				t.Errorf("BodyText = %q, want %q", got.BodyText, tt.want.BodyText)
			}
			if !slices.Equal(got.Links, tt.want.Links) {
				t.Errorf("Links = %v, want %v", got.Links, tt.want.Links)
			}
			if !slices.Equal(got.ImageURLs, tt.want.ImageURLs) {
				t.Errorf("ImageURLs = %v, want %v", got.ImageURLs, tt.want.ImageURLs)
			}
		})
	}
}
