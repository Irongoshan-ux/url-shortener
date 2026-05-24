package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/google/uuid"
)

type fileRecord struct {
	UUID        uuid.UUID `json:"uuid"`
	ShortURL    string    `json:"short_url"`
	OriginalURL string    `json:"original_url"`
	UserID      string    `json:"user_id"`
	IsDeleted   bool      `json:"is_deleted"`
}

// FileRepository persists data to a JSON file while delegating to an embedded MemoryRepository.
type FileRepository struct {
	path string
	mem  *MemoryRepository
	mu   sync.RWMutex
}

// NewFileRepository loads existing records from path when non-empty; path "" yields an in-memory-only store.
func NewFileRepository(path string) (*FileRepository, error) {
	mem := NewMemoryRepository()
	f := &FileRepository{path: path, mem: mem}
	if path == "" {
		return f, nil
	}
	if err := f.load(); err != nil {
		return nil, fmt.Errorf("load storage file: %w", err)
	}
	return f, nil
}

func (f *FileRepository) load() error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var records []fileRecord
	if err := json.Unmarshal(data, &records); err != nil {
		return fmt.Errorf("decode storage file: %w", err)
	}
	for _, r := range records {
		url := &model.URL{
			OriginalURL: r.OriginalURL,
			ShortURL:    r.ShortURL,
			UserID:      r.UserID,
			IsDeleted:   r.IsDeleted,
		}
		_ = f.mem.Create(context.Background(), url)
	}
	return nil
}

func (f *FileRepository) save() error {
	if f.path == "" {
		return nil
	}
	f.mem.mu.RLock()
	records := make([]fileRecord, 0, len(f.mem.urlsByShort))
	for _, u := range f.mem.urlsByShort {
		records = append(records, fileRecord{
			UUID:        uuid.New(),
			ShortURL:    u.ShortURL,
			OriginalURL: u.OriginalURL,
			UserID:      u.UserID,
			IsDeleted:   u.IsDeleted,
		})
	}
	f.mem.mu.RUnlock()

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(f.path, data, 0644)
}

func (f *FileRepository) Create(ctx context.Context, url *model.URL) error {
	if err := f.mem.Create(ctx, url); err != nil {
		return err
	}
	f.mu.Lock()
	err := f.save()
	f.mu.Unlock()
	return err
}

func (f *FileRepository) CreateBatch(ctx context.Context, urls []*model.URL) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.mem.CreateBatch(ctx, urls); err != nil {
		return err
	}
	return f.save()
}

func (f *FileRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	return f.mem.GetByShortURL(ctx, shortURL)
}

func (f *FileRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	return f.mem.GetByOriginalURL(ctx, originalURL)
}

func (f *FileRepository) GetByUserID(ctx context.Context, userID string) ([]*model.URL, error) {
	return f.mem.GetByUserID(ctx, userID)
}

func (f *FileRepository) DeleteByShortURLs(ctx context.Context, userID string, shortIDs []string) error {
	if err := f.mem.DeleteByShortURLs(ctx, userID, shortIDs); err != nil {
		return err
	}
	f.mu.Lock()
	err := f.save()
	f.mu.Unlock()
	return err
}
