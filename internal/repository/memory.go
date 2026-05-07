package repository

import (
	"context"
	"sync"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

type MemoryRepository struct {
	mu             sync.RWMutex
	urlsByShort    map[string]*model.URL
	urlsByOriginal map[string][]*model.URL
	urlsByUser     map[string][]*model.URL
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		urlsByShort:    make(map[string]*model.URL),
		urlsByOriginal: make(map[string][]*model.URL),
		urlsByUser:     make(map[string][]*model.URL),
	}
}

func (r *MemoryRepository) Create(ctx context.Context, url *model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, u := range r.urlsByOriginal[url.OriginalURL] {
		if !u.IsDeleted {
			return ErrConflict
		}
	}
	if _, exists := r.urlsByShort[url.ShortURL]; exists {
		return ErrAlreadyExists
	}

	r.urlsByShort[url.ShortURL] = url
	r.urlsByOriginal[url.OriginalURL] = append(r.urlsByOriginal[url.OriginalURL], url)
	if url.UserID != "" {
		r.urlsByUser[url.UserID] = append(r.urlsByUser[url.UserID], url)
	}

	return nil
}

func (r *MemoryRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, u := range urls {
		for _, existing := range r.urlsByOriginal[u.OriginalURL] {
			if !existing.IsDeleted {
				return ErrConflict
			}
		}
		if _, exists := r.urlsByShort[u.ShortURL]; exists {
			return ErrAlreadyExists
		}
	}
	for _, u := range urls {
		r.urlsByShort[u.ShortURL] = u
		r.urlsByOriginal[u.OriginalURL] = append(r.urlsByOriginal[u.OriginalURL], u)
		if u.UserID != "" {
			r.urlsByUser[u.UserID] = append(r.urlsByUser[u.UserID], u)
		}
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

	list := r.urlsByOriginal[originalURL]
	for _, u := range list {
		if !u.IsDeleted {
			return u, nil
		}
	}
	return nil, ErrNotFound
}

func (r *MemoryRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := r.urlsByUser[userID]
	if len(list) == 0 {
		return nil, nil
	}
	out := make([]*model.URL, 0, len(list))
	for _, u := range list {
		if !u.IsDeleted {
			out = append(out, u)
		}
	}
	return out, nil
}

func (r *MemoryRepository) DeleteByShortURLs(ctx context.Context, userID string, shortIDs []string) error {
	if len(shortIDs) == 0 {
		return nil
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, id := range shortIDs {
		u, exists := r.urlsByShort[id]
		if exists && u.UserID == userID {
			u.IsDeleted = true
		}
	}
	return nil
}
