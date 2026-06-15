// Package service implements URL shortening, batch creation, lookup, listing by user, and asynchronous soft-delete against a Repository.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
)

// ErrConflict is returned by ShortenURL when the original URL is already stored (non-deleted); the existing *model.URL is returned with the error for handlers to respond with 409.
var ErrConflict = errors.New("url already shortened")

const deleteWorkerBatchDelay = 100 * time.Millisecond
const deleteWorkerMaxBatch = 100
const deleteWriterSemSize = 100

type deleteJob struct {
	userID   string
	shortIDs []string
}

// Service coordinates shortening, redirects data access through Repository, and batches delete operations in a background worker.
type Service struct {
	repo         Repository
	maxAttempts  int
	idGen        func() (string, error)
	deleteCh     chan deleteJob
	deleteSem    chan struct{} // limits concurrent goroutines sending to deleteCh
	deleteWorker sync.Once
}

// NewService starts the delete worker goroutine. Apply WithIDGenerator or WithMaxAttempts to customize behavior (tests, collision handling).
func NewService(repo Repository, opts ...Option) *Service {
	s := &Service{
		repo:        repo,
		maxAttempts: 10,
		deleteCh:    make(chan deleteJob, 1024),
		deleteSem:   make(chan struct{}, deleteWriterSemSize),
	}
	s.idGen = s.generateShortID
	for _, opt := range opts {
		opt(s)
	}
	s.startDeleteWorker()
	return s
}

func (s *Service) startDeleteWorker() {
	s.deleteWorker.Do(func() {
		go s.runDeleteWorker()
	})
}

func (s *Service) runDeleteWorker() {
	ticker := time.NewTicker(deleteWorkerBatchDelay)
	defer ticker.Stop()
	var batch []deleteJob
	flush := func() {
		if len(batch) == 0 {
			return
		}
		byUser := make(map[string]map[string]bool)
		for _, j := range batch {
			if byUser[j.userID] == nil {
				byUser[j.userID] = make(map[string]bool)
			}
			for _, id := range j.shortIDs {
				byUser[j.userID][id] = true
			}
		}
		ctx := context.Background()
		for userID, idsSet := range byUser {
			shortIDs := make([]string, 0, len(idsSet))
			for id := range idsSet {
				shortIDs = append(shortIDs, id)
			}
			_ = s.repo.DeleteByShortURLs(ctx, userID, shortIDs)
		}
		batch = batch[:0]
	}
	for {
		select {
		case j, ok := <-s.deleteCh:
			if !ok {
				flush()
				return
			}
			batch = append(batch, j)
			if len(batch) >= deleteWorkerMaxBatch {
				flush()
			}
		case <-ticker.C:
			flush()
		}
	}
}

// ShortenURL persists a new short mapping or returns ErrConflict with the existing record if originalURL is already active.
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
					if errors.Is(getErr, repository.ErrNotFound) {
						continue
					}
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

// GetOriginalURL resolves a short id to the stored original URL string.
func (s *Service) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	url, err := s.repo.GetByShortURL(ctx, shortID)
	if err != nil {
		return "", fmt.Errorf("failed to get URL: %w", err)
	}
	return url.OriginalURL, nil
}

// GetURLByShortID returns the full URL record, including deletion flag for HTTP 410 handling.
func (s *Service) GetURLByShortID(ctx context.Context, shortID string) (*model.URL, error) {
	return s.repo.GetByShortURL(ctx, shortID)
}

// DeleteUserURLs enqueues soft-deletes for the given short ids; it returns before the repository is updated.
func (s *Service) DeleteUserURLs(ctx context.Context, userID string, shortIDs []string) {
	if len(shortIDs) == 0 {
		return
	}
	job := deleteJob{userID: userID, shortIDs: shortIDs}
	go func() {
		s.deleteSem <- struct{}{}
		defer func() { <-s.deleteSem }()
		s.deleteCh <- job
	}()
}

// BatchItem is one entry in a batch shorten request (correlation id is echoed in the response).
type BatchItem struct {
	CorrelationID string
	OriginalURL   string
}

// BatchResult maps a correlation id to the generated or existing short id (not a full URL).
type BatchResult struct {
	CorrelationID string
	ShortURL      string
}

// ShortenURLBatch creates missing mappings in one pass and fills results in request order.
func (s *Service) ShortenURLBatch(ctx context.Context, items []BatchItem, userID string) ([]BatchResult, error) {
	if len(items) == 0 {
		return nil, nil
	}
	results := make([]BatchResult, len(items))
	type indexCorr struct {
		index         int
		correlationID string
	}
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

// GetUserURLs lists all non-deleted URLs owned by userID.
func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]*model.URL, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// GetStats returns counts of active shortened URLs and distinct users.
func (s *Service) GetStats(ctx context.Context) (urls, users int, err error) {
	urls, err = s.repo.CountURLs(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("count urls: %w", err)
	}
	users, err = s.repo.CountUsers(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("count users: %w", err)
	}
	return urls, users, nil
}

func (s *Service) generateShortID() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	var enc [8]byte
	base64.RawURLEncoding.Encode(enc[:], b)
	return string(enc[:]), nil
}
