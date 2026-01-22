package validation

import (
	"fmt"
	"net/url"
	"strings"
)

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

func NormalizeBaseURL(baseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	parsed, err := ParseHTTPURL(baseURL)
	if err != nil {
		return "", err
	}
	return parsed.String(), nil
}


