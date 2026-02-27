package repository

import (
	"context"
	"fmt"
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
		return fmt.Errorf("short id %q already exists: %w", url.ShortURL, ErrAlreadyExists)
	}

	if existing, exists := r.urlsByOriginal[url.OriginalURL]; exists {
		return fmt.Errorf("original url %q already shortened as %q: %w", url.OriginalURL, existing.ShortURL, ErrAlreadyExists)
	}

	r.urlsByShort[url.ShortURL] = url
	r.urlsByOriginal[url.OriginalURL] = url

	return nil
}

func (r *MemoryRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	if len(urls) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range urls {
		if _, exists := r.urlsByShort[u.ShortURL]; exists {
			return fmt.Errorf("short id %q already exists: %w", u.ShortURL, ErrAlreadyExists)
		}
		if _, exists := r.urlsByOriginal[u.OriginalURL]; exists {
			return fmt.Errorf("original url %q already in batch: %w", u.OriginalURL, ErrAlreadyExists)
		}
	}
	for _, u := range urls {
		r.urlsByShort[u.ShortURL] = u
		r.urlsByOriginal[u.OriginalURL] = u
	}
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
