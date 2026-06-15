package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"github.com/Irongoshan-ux/url-shortener/internal/handler/mocks"
	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"go.uber.org/mock/gomock"
)

func withBenchUser(ctx context.Context) context.Context {
	ctx = auth.WithUserID(ctx, "bench-user")
	ctx = auth.WithHadValidCookie(ctx, true)
	ctx = auth.WithHadCookieInRequest(ctx, true)
	return ctx
}

func BenchmarkHandler_ShortenURLJSON(b *testing.B) {
	ctrl := gomock.NewController(b)
	svc := mocks.NewMockURLService(ctrl)
	svc.EXPECT().
		ShortenURL(gomock.Any(), "https://example.com/page", gomock.Any()).
		Return(&model.URL{OriginalURL: "https://example.com/page", ShortURL: "shortid1", CreatedAt: time.Now()}, nil).
		AnyTimes()

	h := NewHandler(svc, "http://localhost:8080", zerolog.Nop(), nil, "")
	body := `{"url":"https://example.com/page"}`

	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
		req = req.WithContext(withBenchUser(req.Context()))
		req.Header.Set("Content-Type", "application/json")
		req.Host = "localhost:8080"
		w := httptest.NewRecorder()
		b.StartTimer()
		h.ShortenURLJSON(w, req)
	}
}

func BenchmarkHandler_Redirect(b *testing.B) {
	ctrl := gomock.NewController(b)
	svc := mocks.NewMockURLService(ctrl)
	svc.EXPECT().
		GetURLByShortID(gomock.Any(), "abc123").
		Return(&model.URL{OriginalURL: "https://target.example/foo", ShortURL: "abc123", IsDeleted: false}, nil).
		AnyTimes()

	h := NewHandler(svc, "", zerolog.Nop(), nil, "")
	r := chi.NewRouter()
	r.Get("/{id}", h.Redirect)

	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/abc123", nil)
		req = req.WithContext(withBenchUser(req.Context()))
		w := httptest.NewRecorder()
		b.StartTimer()
		r.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ParseAndShorten_plain(b *testing.B) {
	repo := repository.NewMemoryRepository()
	svc := service.NewService(repo)
	h := NewHandler(svc, "http://localhost:8080", zerolog.Nop(), nil, "")
	r := chi.NewRouter()
	r.Use(auth.CookieMiddleware("s"))
	r.Mount("/", h.Router())

	var n int
	b.ReportAllocs()
	for b.Loop() {
		b.StopTimer()
		body := "https://plain.example/post/" + strings.Repeat("x", n%20)
		n++
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
		req.Host = "localhost:8080"
		w := httptest.NewRecorder()
		b.StartTimer()
		r.ServeHTTP(w, req)
	}
}
