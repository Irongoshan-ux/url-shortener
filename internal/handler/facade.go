package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
	"github.com/rs/zerolog"
)

var (
	// ErrEmptyURL is returned when the original URL is missing.
	ErrEmptyURL = errors.New("url is required")
	// ErrInvalidURL is returned when URL validation fails.
	ErrInvalidURL = errors.New("invalid url")
	// ErrExpandNotFound is returned when a short id does not exist.
	ErrExpandNotFound = errors.New("url not found")
	// ErrExpandGone is returned when a short id refers to a soft-deleted URL.
	ErrExpandGone = errors.New("url gone")
)

// ShortenerFacade contains transport-agnostic URL shortener operations.
type ShortenerFacade struct {
	service URLService
	baseURL string
	log     zerolog.Logger
}

// NewShortenerFacade constructs a facade over URLService.
func NewShortenerFacade(svc URLService, baseURL string, log zerolog.Logger) *ShortenerFacade {
	return &ShortenerFacade{
		service: svc,
		baseURL: baseURL,
		log:     log,
	}
}

func (f *ShortenerFacade) buildFullURL(r *http.Request, shortID string) (string, error) {
	if f.baseURL != "" {
		var b strings.Builder
		b.Grow(len(f.baseURL) + 1 + len(shortID))
		b.WriteString(f.baseURL)
		b.WriteByte('/')
		b.WriteString(shortID)
		return b.String(), nil
	}
	if r != nil {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		base := scheme + "://" + r.Host
		return url.JoinPath(base, shortID)
	}
	return shortID, nil
}

// ShortenURL validates and shortens originalURL for userID. Returns the short id.
func (f *ShortenerFacade) ShortenURL(ctx context.Context, originalURL, userID string) (shortID string, conflict bool, err error) {
	originalURL = strings.TrimSpace(originalURL)
	if originalURL == "" {
		return "", false, ErrEmptyURL
	}
	parsedURL, err := validation.ParseHTTPURL(originalURL)
	if err != nil {
		return "", false, fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}
	normalizedURL := parsedURL.String()

	shortURL, err := f.service.ShortenURL(ctx, normalizedURL, userID)
	if err != nil {
		if errors.Is(err, service.ErrConflict) && shortURL != nil {
			return shortURL.ShortURL, true, nil
		}
		return "", false, err
	}
	return shortURL.ShortURL, false, nil
}

// ExpandURL resolves shortID to the original URL.
func (f *ShortenerFacade) ExpandURL(ctx context.Context, shortID string) (string, error) {
	shortID = strings.TrimSpace(shortID)
	if shortID == "" {
		return "", ErrExpandNotFound
	}
	url, err := f.service.GetURLByShortID(ctx, shortID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrExpandNotFound
		}
		return "", err
	}
	if url.IsDeleted {
		return "", ErrExpandGone
	}
	return url.OriginalURL, nil
}

// ListUserURLs returns active URLs for userID. When r is non-nil and baseURL is empty, absolute short links use the request host.
func (f *ShortenerFacade) ListUserURLs(ctx context.Context, userID string, r *http.Request) ([]userURLItem, error) {
	urls, err := f.service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(urls) == 0 {
		return nil, nil
	}
	resp := make([]userURLItem, len(urls))
	for i, u := range urls {
		fullShort, _ := f.buildFullURL(r, u.ShortURL)
		resp[i] = userURLItem{
			ShortURL:    fullShort,
			OriginalURL: u.OriginalURL,
		}
	}
	return resp, nil
}

// BuildFullURLFromBase builds an absolute short link using configured baseURL.
func (f *ShortenerFacade) BuildFullURLFromBase(shortID string) string {
	if f.baseURL == "" {
		return shortID
	}
	var b strings.Builder
	b.Grow(len(f.baseURL) + 1 + len(shortID))
	b.WriteString(f.baseURL)
	b.WriteByte('/')
	b.WriteString(shortID)
	return b.String()
}

// GetStats returns service statistics.
func (f *ShortenerFacade) GetStats(ctx context.Context) (urls, users int, err error) {
	return f.service.GetStats(ctx)
}
