package integrationbench

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"github.com/Irongoshan-ux/url-shortener/internal/handler"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
)

func BenchmarkWorkloadMixed(b *testing.B) {
	repo := repository.NewMemoryRepository()
	svc := service.NewService(repo)
	h := handler.NewHandler(handler.NewShortenerFacade(svc, "http://bench.example", zerolog.Nop()), zerolog.Nop(), nil, "")
	r := chi.NewRouter()
	r.Use(auth.CookieMiddleware("bench-secret-key"))
	r.Mount("/", h.Router())

	ctx := context.Background()
	shortIDs := make([]string, 0, 64)
	for i := range 64 {
		u, err := svc.ShortenURL(ctx, fmt.Sprintf("https://seed.example/path/%d?q=%d", i, i), "bench-user")
		if err != nil {
			b.Fatal(err)
		}
		shortIDs = append(shortIDs, u.ShortURL)
	}

	j := 0
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%3 == 0 {
			body := `{"url":"https://load.test/item/` + strconv.Itoa(i) + `"}`
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req.Host = "bench.example"
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			continue
		}
		id := shortIDs[j%len(shortIDs)]
		j++
		req := httptest.NewRequest(http.MethodGet, "/"+id, nil)
		req.Host = "bench.example"
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
