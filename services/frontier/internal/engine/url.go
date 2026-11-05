package engine

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// NormalizeURL canonicalizes a URL string for crawler deduplication.
func NormalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", fmt.Errorf("parse URL: %w", err)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", errors.New("invalid URL scheme: must be http or https")
	}
	u.Scheme = scheme

	if strings.TrimSpace(u.Host) == "" {
		return "", errors.New("url host cannot be empty")
	}

	port := u.Port()
	if (scheme == "http" && port == "80") || (scheme == "https" && port == "443") {
		u.Host = strings.ToLower(u.Hostname())
	} else {
		u.Host = strings.ToLower(u.Host)
	}

	u.Fragment = ""

	u.RawQuery = u.Query().Encode()

	if u.Path == "" || u.Path == "/" {
		u.Path = "/"
	} else {
		u.Path = strings.TrimSuffix(u.Path, "/")
	}

	return u.String(), nil
}
