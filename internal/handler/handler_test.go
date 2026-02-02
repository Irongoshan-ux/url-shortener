package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/handler/mocks"
	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/go-chi/chi/v5"
	"go.uber.org/mock/gomock"
)

func TestHandler_ShortenURL(t *testing.T) {
	tests := []struct {
		name            string
		method          string
		body            string
		expectedStatus  int
		expectedBody    string
		expectedOrigURL string
	}{
		{
			name:   "successful shorten",
			method: http.MethodPost,
			body:   "https://example.com",
			expectedStatus:  http.StatusCreated,
			expectedOrigURL: "https://example.com",
		},
		{
			name:   "empty body",
			method: http.MethodPost,
			body:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid URL format",
			method: http.MethodPost,
			body:   "not a url",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "URL without scheme",
			method: http.MethodPost,
			body:   "example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "URL without host",
			method: http.MethodPost,
			body:   "https://",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid scheme",
			method: http.MethodPost,
			body:   "ftp://example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "service error",
			method: http.MethodPost,
			body:   "https://example.com",
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "service returns deterministic short id",
			method: http.MethodPost,
			body:   "https://example.com",
			expectedStatus: http.StatusCreated,
			expectedBody:   "http://localhost/existing123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockURLService(ctrl)

			// Expectations for cases where handler should call service.
			switch tt.name {
			case "successful shorten":
				svc.EXPECT().
					ShortenURL(gomock.Any(), "https://example.com").
					Return(&model.URL{OriginalURL: "https://example.com", ShortURL: "abc123", CreatedAt: time.Now()}, nil)
			case "service error":
				svc.EXPECT().
					ShortenURL(gomock.Any(), "https://example.com").
					Return(nil, errors.New("service failure"))
			case "service returns deterministic short id":
				svc.EXPECT().
					ShortenURL(gomock.Any(), "https://example.com").
					Return(&model.URL{OriginalURL: "https://example.com", ShortURL: "existing123", CreatedAt: time.Now()}, nil)
			default:
			}

			h := NewHandler(svc, "")

			req := httptest.NewRequest(tt.method, "/", strings.NewReader(tt.body))
			req.Host = "localhost"
			w := httptest.NewRecorder()

			h.ShortenURL(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
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
		expectedStatus   int
		expectedLocation string
	}{
		{
			name:   "successful redirect",
			method: http.MethodGet,
			path:   "abc123",
			expectedStatus:   http.StatusTemporaryRedirect,
			expectedLocation: "https://example.com",
		},
		{
			name:   "URL not found",
			method: http.MethodGet,
			path:   "nonexistent",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "wrong method",
			method: http.MethodPost,
			path:   "abc123",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockURLService(ctrl)

			switch tt.name {
			case "successful redirect":
				svc.EXPECT().GetOriginalURL(gomock.Any(), "abc123").Return("https://example.com", nil)
			case "URL not found":
				svc.EXPECT().GetOriginalURL(gomock.Any(), "nonexistent").Return("", repository.ErrNotFound)
			default:
			}

			h := NewHandler(svc, "")

			r := chi.NewRouter()
			r.Mount("/", h.Router())

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

func TestHandler_ShortenURLJSON(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		baseURL        string
		expectedStatus int
		expectedBody   string
		expectJSON     bool
		setupMock      func(*mocks.MockURLService)
	}{
		{
			name:           "successful shorten",
			body:           `{"url":"https://practicum.yandex.ru"}`,
			baseURL:        "http://localhost:8080",
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"result":"http://localhost:8080/EwHXdJfB"}`,
			expectJSON:     true,
			setupMock: func(s *mocks.MockURLService) {
				s.EXPECT().
					ShortenURL(gomock.Any(), "https://practicum.yandex.ru").
					Return(&model.URL{OriginalURL: "https://practicum.yandex.ru", ShortURL: "EwHXdJfB", CreatedAt: time.Now()}, nil)
			},
		},
		{
			name:           "empty url",
			body:           `{"url":""}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(*mocks.MockURLService) {},
		},
		{
			name:           "missing url field",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(*mocks.MockURLService) {},
		},
		{
			name:           "invalid JSON",
			body:           `{url: no quotes}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(*mocks.MockURLService) {},
		},
		{
			name:           "invalid URL format",
			body:           `{"url":"not a url"}`,
			expectedStatus: http.StatusBadRequest,
			setupMock:      func(*mocks.MockURLService) {},
		},
		{
			name:           "service error",
			body:           `{"url":"https://example.com"}`,
			expectedStatus: http.StatusInternalServerError,
			setupMock: func(s *mocks.MockURLService) {
				s.EXPECT().
					ShortenURL(gomock.Any(), "https://example.com").
					Return(nil, errors.New("service failure"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockURLService(ctrl)
			tt.setupMock(svc)

			h := NewHandler(svc, tt.baseURL)

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Host = "localhost"
			w := httptest.NewRecorder()

			r := chi.NewRouter()
			r.Mount("/", h.Router())
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectJSON {
				if ct := w.Header().Get("Content-Type"); ct != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", ct)
				}
				body := strings.TrimSpace(w.Body.String())
				if body != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, body)
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
		expectedStatus int
	}{
		{
			name:   "GET / - should return error",
			method: http.MethodGet,
			path:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /smth - should return error",
			method: http.MethodGet,
			path:   "smth",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /smth/else - should return error",
			method: http.MethodGet,
			path:   "smth/else",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "GET /{id} - should redirect",
			method: http.MethodGet,
			path:   "abc123",
			expectedStatus: http.StatusTemporaryRedirect,
		},
		{
			name:   "POST / - should shorten",
			method: http.MethodPost,
			path:   "",
			body:   "https://example.com",
			expectedStatus: http.StatusCreated,
		},
		{
			name:   "POST /{id} - should return error",
			method: http.MethodPost,
			path:   "abc123",
			body:   "https://example.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST / without body - should return error",
			method: http.MethodPost,
			path:   "",
			body:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "POST /smth - should return error",
			method: http.MethodPost,
			path:   "smth",
			body:   "http://test.com",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "PUT / - should return error",
			method: http.MethodPut,
			path:   "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "POST /api/shorten - should shorten",
			method:         http.MethodPost,
			path:           "api/shorten",
			body:           `{"url":"https://example.com"}`,
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := mocks.NewMockURLService(ctrl)
			h := NewHandler(svc, "")

			r := chi.NewRouter()
			r.Mount("/", h.Router())

			switch tt.name {
			case "GET /smth - should return error":
				svc.EXPECT().GetOriginalURL(gomock.Any(), "smth").Return("", repository.ErrNotFound)
			case "GET /{id} - should redirect":
				svc.EXPECT().GetOriginalURL(gomock.Any(), "abc123").Return("https://example.com", nil)
			case "POST / - should shorten":
				svc.EXPECT().ShortenURL(gomock.Any(), "https://example.com").Return(&model.URL{OriginalURL: "https://example.com", ShortURL: "abc123", CreatedAt: time.Now()}, nil)
			case "POST /api/shorten - should shorten":
				svc.EXPECT().ShortenURL(gomock.Any(), "https://example.com").Return(&model.URL{OriginalURL: "https://example.com", ShortURL: "abc123", CreatedAt: time.Now()}, nil)
			default:
			}

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
