package handler

import (
	"context"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
)

// URLService is the facade HTTP handlers use; *service.Service satisfies this interface.
//
//go:generate go run go.uber.org/mock/mockgen -source=service.go -destination=mocks/mock_urlservice.go -package=mocks
type URLService interface {
	// ShortenURL persists a new short link or surfaces an existing one with service.ErrConflict.
	ShortenURL(ctx context.Context, originalURL string, userID string) (*model.URL, error)
	// ShortenURLBatch creates many links in one repository round-trip when possible.
	ShortenURLBatch(ctx context.Context, items []service.BatchItem, userID string) ([]service.BatchResult, error)
	// GetOriginalURL resolves shortID to the original URL string.
	GetOriginalURL(ctx context.Context, shortID string) (string, error)
	// GetURLByShortID loads the full record, including deletion state.
	GetURLByShortID(ctx context.Context, shortID string) (*model.URL, error)
	// GetUserURLs returns URLs created for the given user id.
	GetUserURLs(ctx context.Context, userID string) ([]*model.URL, error)
	// DeleteUserURLs enqueues soft-deletes for the listed short ids.
	DeleteUserURLs(ctx context.Context, userID string, shortIDs []string)
}
