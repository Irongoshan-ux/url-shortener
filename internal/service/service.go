package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
)

type Service struct {
	repo Repository
	maxAttempts int
	idGen       func() (string, error)
}

func NewService(repo Repository, opts ...Option) *Service {
	s := &Service{
		repo:        repo,
		maxAttempts: 10,
	}
	s.idGen = s.generateShortID
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) ShortenURL(ctx context.Context, originalURL string) (*model.URL, error) {
	existingURL, err := s.repo.GetByOriginalURL(ctx, originalURL)
	if err == nil {
		return existingURL, nil
	}
	if err != repository.ErrNotFound {
		return nil, fmt.Errorf("failed to check existing URL: %w", err)
	}

	for attempt := 0; attempt < s.maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		shortID, err := s.idGen()
		if err != nil {
			return nil, fmt.Errorf("failed to generate short ID: %w", err)
		}

		url := &model.URL{
			OriginalURL: originalURL,
			ShortURL:    shortID,
			CreatedAt:   time.Now(),
		}

		if err := s.repo.Create(ctx, url); err != nil {
			if !errors.Is(err, repository.ErrAlreadyExists) {
				return nil, fmt.Errorf("failed to create URL: %w", err)
			}

			if existingURL, getErr := s.repo.GetByOriginalURL(ctx, originalURL); getErr == nil {
				return existingURL, nil
			}

			continue
		}

		return url, nil
	}

	return nil, fmt.Errorf("failed to create URL after %d attempts due to short id collisions", s.maxAttempts)
}

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
