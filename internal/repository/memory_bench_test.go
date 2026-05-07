package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
)

func BenchmarkMemoryRepository_Create(b *testing.B) {
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		r := NewMemoryRepository()
		_ = r.Create(ctx, &model.URL{
			OriginalURL: "https://example.com/page",
			ShortURL:    "short01",
			CreatedAt:   time.Now(),
			UserID:      "u1",
		})
	}
}

func BenchmarkMemoryRepository_GetByShortURL(b *testing.B) {
	r := NewMemoryRepository()
	_ = r.Create(context.Background(), &model.URL{
		OriginalURL: "https://example.com/p",
		ShortURL:    "abc12345",
		CreatedAt:   time.Now(),
		UserID:      "u1",
	})
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = r.GetByShortURL(ctx, "abc12345")
	}
}

func BenchmarkMemoryRepository_GetByOriginalURL(b *testing.B) {
	r := NewMemoryRepository()
	const orig = "https://example.com/original"
	_ = r.Create(context.Background(), &model.URL{
		OriginalURL: orig,
		ShortURL:    "id999999",
		CreatedAt:   time.Now(),
		UserID:      "u1",
	})
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = r.GetByOriginalURL(ctx, orig)
	}
}
