// Package validation provides strict parsing for user-supplied and configured HTTP(S) URLs.
package validation

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseHTTPURL accepts only http/https URLs with a non-empty host.
func ParseHTTPURL(raw string) (*url.URL, error) {
	raw = strings.TrimSpace(raw)
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	if parsed.Scheme == "" {
		return nil, fmt.Errorf("url must include scheme (http:// or https://)")
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("url must include host")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, fmt.Errorf("url scheme must be http or https")
	}
	return parsed, nil
}

// NormalizeBaseURL trims space and trailing slashes, validates, and returns canonical string for building short links.
func NormalizeBaseURL(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := ParseHTTPURL(baseURL)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}
