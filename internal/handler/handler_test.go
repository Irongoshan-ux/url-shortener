package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

func newTestHandler(svc *service.Service) *Handler {
	return NewHandler(svc, "")
}

type mockRepository struct {
	urlsByShort    map[string]*model.URL
	urlsByOriginal map[string]*model.URL
	createError    error
	getError       error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		urlsByShort:    make(map[string]*model.URL),
		urlsByOriginal: make(map[string]*model.URL),
	}
}

func (m *mockRepository) Create(ctx context.Context, url *model.URL) error {
	if m.createError != nil {
		return m.createError
	}
	if _, exists := m.urlsByShort[url.ShortURL]; exists {
		return repository.ErrAlreadyExists
	}
	if _, exists := m.urlsByOriginal[url.OriginalURL]; exists {
		return repository.ErrAlreadyExists
	}
	m.urlsByShort[url.ShortURL] = url
	m.urlsByOriginal[url.OriginalURL] = url
	return nil
}

func (m *mockRepository) GetByShortURL(ctx context.Context, shortURL string) (*model.URL, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	url, exists := m.urlsByShort[shortURL]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return url, nil
}

func (m *mockRepository) GetByOriginalURL(ctx context.Context, originalURL string) (*model.URL, error) {
	url, exists := m.urlsByOriginal[originalURL]
	if !exists {
		return nil, repository.ErrNotFound
	}
	return url, nil
}

func TestHandler_ShortenURL(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		body            string
		setupRepo       func() *mockRepository
		expectedStatus  int
		expectedBody    string
		expectedOrigURL string
	}{
		{
			name:   "successful shorten",
			method: http.MethodPost,
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus:  http.StatusCreated,
			expectedOrigURL: "https://example.com",
		},
		{
			name:   "wrong method",
			method: http.MethodGet,
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "empty body",
			method: http.MethodPost,
			body:   "",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid URL format",
			method: http.MethodPost,
			body:   "not a url",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "URL without scheme",
			method: http.MethodPost,
			body:   "example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "URL without host",
			method: http.MethodPost,
			body:   "https://",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid scheme",
			method: http.MethodPost,
			body:   "ftp://example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "repository create error",
			method: http.MethodPost,
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				repo := newMockRepository()
				repo.createError = errors.New("repository error")
				return repo
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "existing URL returns same short URL",
			method: http.MethodPost,
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				repo := newMockRepository()
				repo.urlsByOriginal["https://example.com"] = &model.URL{
					OriginalURL: "https://example.com",
					ShortURL:    "existing123",
					CreatedAt:   time.Now(),
				}
				repo.urlsByShort["existing123"] = repo.urlsByOriginal["https://example.com"]
				return repo
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost/existing123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := service.NewService(repo)
			h := newTestHandler(svc)

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Host = "localhost"
			w := httptest.NewRecorder()

			h.ShortenURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedOrigURL != "" && tt.expectedStatus == http.StatusCreated {
				storedURL, err := repo.GetByOriginalURL(req.Context(), tt.expectedOrigURL)
				if err != nil {
					t.Errorf("expected original URL %q to be stored, but got error: %v", tt.expectedOrigURL, err)
				} else if storedURL.OriginalURL != tt.expectedOrigURL {
					t.Errorf("expected stored original URL %q, got %q", tt.expectedOrigURL, storedURL.OriginalURL)
				}
			}

			if tt.expectedBody != "" {
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
				}
			}

			if tt.expectedStatus == http.StatusCreated {
				if w.Header().Get("Content-Type") != "text/plain" {
					t.Errorf("expected Content-Type text/plain, got %s", w.Header().Get("Content-Type"))
				}
			}
		})
	}
}

func TestHandler_Redirect(t *testing.T) {
	tests := []struct {
		name             string
		method           string
		path             string
		setupRepo        func() *mockRepository
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:   "successful redirect",
			method: http.MethodGet,
			path:   "abc123",
			setupRepo: func() *mockRepository {
				repo := newMockRepository()
				repo.urlsByShort["abc123"] = &model.URL{
					OriginalURL: "https://example.com",
					ShortURL:    "abc123",
					CreatedAt:   time.Now(),
				}
				return repo
			},
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:   "URL not found",
			method: http.MethodGet,
			path:   "nonexistent",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			path:   "abc123",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := service.NewService(repo)
			h := newTestHandler(svc)

			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(tt.method, "/"+tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedLocation != "" {
				location := w.Header().Get("Location")
				if location != tt.expectedLocation {
					t.Errorf("expected Location %q, got %q", tt.expectedLocation, location)
				}
			}
		})
	}
}

func TestHandler_Root(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		setupRepo      func() *mockRepository
		expectedStatus int
	}{
		{
			name:   "GET / - should return error",
			method: http.MethodGet,
			path:   "",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /smth - should return error",
			method: http.MethodGet,
			path:   "smth",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /smth/else - should return error",
			method: http.MethodGet,
			path:   "smth/else",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /{id} - should redirect",
			method: http.MethodGet,
			path:   "abc123",
			setupRepo: func() *mockRepository {
				repo := newMockRepository()
				repo.urlsByShort["abc123"] = &model.URL{
					OriginalURL: "https://example.com",
					ShortURL:    "abc123",
					CreatedAt:   time.Now(),
				}
				return repo
			},
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "POST / - should shorten",
			method: http.MethodPost,
			path:   "",
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "POST /{id} - should return error",
			method: http.MethodPost,
			path:   "abc123",
			body:   "https://example.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST / without body - should return error",
			method: http.MethodPost,
			path:   "",
			body:   "",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST /smth - should return error",
			method: http.MethodPost,
			path:   "smth",
			body:   "http://test.com",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "PUT / - should return error",
			method: http.MethodPut,
			path:   "",
			setupRepo: func() *mockRepository {
				return newMockRepository()
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			svc := service.NewService(repo)
			h := newTestHandler(svc)

			r := chi.NewRouter()
			h.RegisterRoutes(r)

			req := httptest.NewRequest(tt.method, "/"+tt.path, strings.NewReader(tt.body))
			req.Host = "localhost"
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
