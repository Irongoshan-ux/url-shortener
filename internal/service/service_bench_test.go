package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
)

func BenchmarkService_ShortenURL(b *testing.B) {
	ctx := context.Background()
	repo := repository.NewMemoryRepository()
	svc := NewService(repo)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = svc.ShortenURL(ctx, fmt.Sprintf("https://bench.example/url/%d", i), "user-1")
	}
}

func BenchmarkService_GetURLByShortID(b *testing.B) {
	repo := repository.NewMemoryRepository()
	sid := "fixedshort"
	_ = repo.Create(context.Background(), &model.URL{
		OriginalURL: "https://example.com/long/path",
		ShortURL:    sid,
		CreatedAt:   time.Now(),
		UserID:      "u",
	})
	svc := NewService(repo)
	ctx := context.Background()
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = svc.GetURLByShortID(ctx, sid)
	}
}
