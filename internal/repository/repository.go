package repository

import (
	"context"
	"errors"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

var (
	ErrNotFound      = errors.New("url not found")
	ErrAlreadyExists = errors.New("url already exists")
)

type Repository interface {
	Create(ctx context.Context, url *model.URL) error

	GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error)

	GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error)
}
