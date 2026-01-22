package handler

import (
	"context"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

// URLService describes the business-logic methods used by HTTP handlers.
//go:generate go run go.uber.org/mock/mockgen -source=service.go -destination=mocks/mock_urlservice.go -package=mocks
type URLService interface {
	ShortenURL(ctx context.Context, originalURL string) (*model.URL, error)
	GetOriginalURL(ctx context.Context, shortID string) (string, error)
}


