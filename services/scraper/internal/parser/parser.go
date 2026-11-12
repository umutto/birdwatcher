// Package parser extracts text, links, and image URLs from HTML documents.
package parser

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// Result holds the structured content and resolved URLs extracted from an HTML document.
type Result struct {
	Title     string
	BodyText  string
	Links     []string
	ImageURLs []string
}

// Parse extracts the page title, visible body text, outgoing links, and image URLs from r,
// resolving relative URLs against baseURL.
func Parse(baseURL string, r io.Reader) (*Result, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, errors.New("invalid base url")
	}

	doc, err := html.Parse(r)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	var bodyBuilder strings.Builder
	seenLinks := make(map[string]struct{})
	seenImages := make(map[string]struct{})
	result := &Result{}

	var walk func(n *html.Node)
	walk = func(n *html.Node) {
		switch n.Type {
		case html.ElementNode:
			switch n.DataAtom {
			case atom.Script, atom.Style, atom.Noscript:
				return
			case atom.Title:
				isPageTitle := result.Title == "" && n.FirstChild != nil && n.FirstChild.Type == html.TextNode
				if isPageTitle {
					result.Title = strings.TrimSpace(n.FirstChild.Data)
				}
				return
			case atom.A:
				if link, ok := resolveURL(base, attr(n, "href")); ok {
					if _, exists := seenLinks[link]; !exists {
						seenLinks[link] = struct{}{}
						result.Links = append(result.Links, link)
					}
				}
			case atom.Img:
				if imgURL, ok := resolveURL(base, attr(n, "src")); ok {
					if _, exists := seenImages[imgURL]; !exists {
						seenImages[imgURL] = struct{}{}
						result.ImageURLs = append(result.ImageURLs, imgURL)
					}
				}
			}
		case html.TextNode:
			_, _ = bodyBuilder.WriteString(n.Data)
			_ = bodyBuilder.WriteByte(' ')
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	result.BodyText = strings.Join(strings.Fields(bodyBuilder.String()), " ")

	return result, nil
}

// attr returns the value of the attribute with the given key, or "" if not found.
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

// resolveURL parses raw, resolves it against base, and ensures it is an HTTP or HTTPS URL without fragments.
func resolveURL(base *url.URL, raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.HasPrefix(raw, "#") {
		return "", false
	}

	ref, err := url.Parse(raw)
	if err != nil {
		return "", false
	}

	resolved := base.ResolveReference(ref)
	if (resolved.Scheme != "http" && resolved.Scheme != "https") || resolved.Host == "" {
		return "", false
	}

	resolved.Fragment = ""
	return resolved.String(), true
}
