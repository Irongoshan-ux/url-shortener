package service

import (
	"context"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

type Repository interface {
	Create(ctx context.Context, url *model.URL) error
	GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error)
	GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error)
}

//go:generate go run go.uber.org/mock/mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks
