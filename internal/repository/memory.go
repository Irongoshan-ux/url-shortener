package repository

import (
	"context"
	"sync"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

type MemoryRepository struct {
	mu             sync.RWMutex
	urlsByShort    map[string]*model.URL
	urlsByOriginal map[string]*model.URL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urlsByShort:    make(map[string]*model.URL),
		urlsByOriginal: make(map[string]*model.URL),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, url *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.urlsByShort[url.ShortURL]; exists {
		return ErrAlreadyExists
	}

	if _, exists := r.urlsByOriginal[url.OriginalURL]; exists {
		return ErrAlreadyExists
	}

	r.urlsByShort[url.ShortURL] = url
	r.urlsByOriginal[url.OriginalURL] = url

	return nil
}

func (r *MemoryRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.urlsByShort[shortURL]
	if !exists {
		return nil, ErrNotFound
	}

	return url, nil
}

func (r *MemoryRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	url, exists := r.urlsByOriginal[originalURL]
	if !exists {
		return nil, ErrNotFound
	}

	return url, nil
}
