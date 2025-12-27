package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (*model.URL, error) {
	// Check if URL already exists
	existingURL, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		return existingURL, nil
	}
	if err != repository.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing URL: %w", err)
	}

	shortID, err := s.generateShortID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate short ID: %w", err)
	}

	url := &model.URL{
		OriginalURL: originalURL,
		ShortURL:    shortID, // Store just the short ID, not the full URL
		CreatedAt:   time.Now(),
	}

	// Store in repository
	if err := s.repo.Create(ctx, url); err != nil {
		return nil, fmt.Errorf("failed to create URL: %w", err)
	}

	return url, nil
}

// GetOriginalURL retrieves the original URL from a short ID
func (s *Service) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	url, err := s.repo.GetByShortURL(ctx, shortID)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}

	return url.OriginalURL, nil
}

func (s *Service) generateShortID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.EncodeToString(b)
	return encoded[:8], nil
}
