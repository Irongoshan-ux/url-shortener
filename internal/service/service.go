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

var ErrConflict = errors.New("url already shortened")

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

func (s *Service) ShortenURL(ctx context.Context, originalURL string, userID string) (*model.URL, error) {
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
			UserID:      userID,
		}

		if err := s.repo.Create(ctx, url); err != nil {
			if errors.Is(err, repository.ErrConflict) {
				existing, getErr := s.repo.GetByOriginalURL(ctx, originalURL)
				if getErr != nil {
					return nil, fmt.Errorf("get existing URL: %w", getErr)
				}
				return existing, ErrConflict
			}
			if !errors.Is(err, repository.ErrAlreadyExists) {
				return nil, fmt.Errorf("failed to create URL: %w", err)
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

type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

type BatchResult struct {
	CorrelationID string
	ShortURL      string
}

func (s *Service) ShortenURLBatch(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, nil
	}
	results := make([]BatchResult, len(items))
	type indexCorr struct{ index int; correlationID string }
	toCreateByURL := make(map[string][]indexCorr)
	var uniqueOrder []string

	for i, item := range items {
		existing, err := s.repo.GetByOriginalURL(ctx, item.OriginalURL)
		if err == nil {
			results[i] = BatchResult{CorrelationID: item.CorrelationID, ShortURL: existing.ShortURL}
			continue
		}
		if err != repository.ErrNotFound {
			return nil, fmt.Errorf("get existing URL: %w", err)
		}
		if _, ok := toCreateByURL[item.OriginalURL]; !ok {
			uniqueOrder = append(uniqueOrder, item.OriginalURL)
		}
		toCreateByURL[item.OriginalURL] = append(toCreateByURL[item.OriginalURL], indexCorr{i, item.CorrelationID})
	}

	if len(uniqueOrder) > 0 {
		urlsToCreate := make([]*model.URL, 0, len(uniqueOrder))
		for _, origURL := range uniqueOrder {
			shortID, err := s.idGen()
			if err != nil {
				return nil, fmt.Errorf("generate short ID: %w", err)
			}
			urlsToCreate = append(urlsToCreate, &model.URL{
				OriginalURL: origURL,
				ShortURL:    shortID,
				CreatedAt:   time.Now(),
				UserID:      userID,
			})
		}
		if err := s.repo.CreateBatch(ctx, urlsToCreate); err != nil {
			return nil, fmt.Errorf("create batch: %w", err)
		}
		for j, u := range urlsToCreate {
			for _, ic := range toCreateByURL[uniqueOrder[j]] {
				results[ic.index] = BatchResult{CorrelationID: ic.correlationID, ShortURL: u.ShortURL}
			}
		}
	}

	return results, nil
}

func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]*model.URL, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) generateShortID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	encoded := base64.URLEncoding.EncodeToString(b)
	return encoded[:8], nil
}
