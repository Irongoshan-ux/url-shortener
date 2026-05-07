package service

import (
	"context"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

// Repository abstracts storage for URL mappings used by Service.
type Repository interface {
	Create(ctx context.Context, url *model.URL) error
	CreateBatch(ctx context.Context, urls []*model.URL) error
	GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error)
	GetByUserID(ctx context.Context, userID string) ([]*model.URL, error)
	DeleteByShortURLs(ctx context.Context, userID string, shortIDs []string) error
}

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks
