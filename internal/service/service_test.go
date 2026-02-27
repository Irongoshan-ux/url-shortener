package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

func TestShortenURL_SameOriginalURL_SecondReturnsConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	idGen := func() (string, error) { return "id1", nil }
	svc := NewService(repo, WithIDGenerator(idGen))

	ctx := context.Background()
	orig := "https://example.com"
	existingURL := &model.URL{OriginalURL: orig, ShortURL: "id1"}

	gomock.InOrder(
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "id1" {
				t.Fatalf("unexpected url: %+v", u)
			}
			return nil
		}),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "id1" {
				t.Fatalf("unexpected url: %+v", u)
			}
			return repository.ErrConflict
		}),
		repo.EXPECT().GetByOriginalURL(gomock.Any(), orig).Return(existingURL, nil),
	)

	u1, err := svc.ShortenURL(ctx, orig)
	if err != nil {
		t.Fatalf("first shorten failed: %v", err)
	}
	if u1.ShortURL != "id1" {
		t.Fatalf("expected short id %q, got %q", "id1", u1.ShortURL)
	}

	u2, err := svc.ShortenURL(ctx, orig)
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("second shorten expected ErrConflict, got err=%v", err)
	}
	if u2 == nil || u2.ShortURL != "id1" {
		t.Fatalf("expected existing short id %q, got %v", "id1", u2)
	}
}

func TestShortenURL_RetriesOnShortIDCollision_UntilSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	ids := []string{"dup", "dup", "ok"}
	genCalls := 0
	idGen := func() (string, error) {
		genCalls++
		return ids[genCalls-1], nil
	}

	svc := NewService(repo, WithIDGenerator(idGen), WithMaxAttempts(10))

	orig := "https://b.example"

	gomock.InOrder(
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "dup" {
				t.Fatalf("unexpected url on attempt1: %+v", u)
			}
			return repository.ErrAlreadyExists
		}),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "dup" {
				t.Fatalf("unexpected url on attempt2: %+v", u)
			}
			return repository.ErrAlreadyExists
		}),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "ok" {
				t.Fatalf("unexpected url on attempt3: %+v", u)
			}
			return nil
		}),
	)

	u, err := svc.ShortenURL(context.Background(), orig)
	if err != nil {
		t.Fatalf("shorten failed: %v", err)
	}
	if u.ShortURL != "ok" {
		t.Fatalf("expected short id %q, got %q", "ok", u.ShortURL)
	}
	if genCalls != 3 {
		t.Fatalf("expected 3 generator calls, got %d", genCalls)
	}
}

func TestShortenURL_MaxAttemptsExceeded_OnPersistentCollision(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mocks.NewMockRepository(ctrl)

	genCalls := 0
	idGen := func() (string, error) {
		genCalls++
		return "dup", nil
	}

	svc := NewService(repo, WithIDGenerator(idGen), WithMaxAttempts(2))

	orig := "https://b.example"

	gomock.InOrder(
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "dup" {
				t.Fatalf("unexpected url on attempt1: %+v", u)
			}
			return repository.ErrAlreadyExists
		}),
		repo.EXPECT().Create(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, u *model.URL) error {
			if u.OriginalURL != orig || u.ShortURL != "dup" {
				t.Fatalf("unexpected url on attempt2: %+v", u)
			}
			return repository.ErrAlreadyExists
		}),
	)

	_, err := svc.ShortenURL(context.Background(), "https://b.example")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	// Ensure it's not context-related error.
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("unexpected context error: %v", err)
	}
	if genCalls != 2 {
		t.Fatalf("expected 2 generator calls, got %d", genCalls)
	}
}
