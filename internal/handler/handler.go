package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *service.Service
	baseURL string
}

func NewHandler(svc *service.Service, baseURL string) *Handler {
	return &Handler{
		service: svc,
		baseURL: baseURL,
	}
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(string(body))
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	if parsedURL.Scheme == "" {
		http.Error(w, "URL must be absolute (include http:// or https://)", http.StatusBadRequest)
		return
	}

	if parsedURL.Host == "" {
		http.Error(w, "URL must include a host", http.StatusBadRequest)
		return
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		http.Error(w, "URL scheme must be http or https", http.StatusBadRequest)
		return
	}

	normalizedURL := parsedURL.String()

	shortURL, err := h.service.ShortenURL(r.Context(), normalizedURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var fullURL string
	if h.baseURL != "" {
		fullURL = h.baseURL + "/" + shortURL.ShortURL
	} else {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		fullURL = scheme + "://" + r.Host + "/" + shortURL.ShortURL
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Short URL ID is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURL(r.Context(), shortID)
	if err != nil {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// POST / - shorten URL
	r.Post("/", h.ShortenURL)

	// GET /{id} - redirect to original URL
	r.Get("/{id}", h.Redirect)

	// GET / - return error
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Short URL ID is required", http.StatusBadRequest)
	})

	// Handle all other methods and paths with 400
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad request", http.StatusBadRequest)
	})

	// Handle 404 (paths with multiple segments like /smth/else)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		// Check if it's a GET request to a path with multiple segments
		if r.Method == http.MethodGet {
			http.Error(w, "Short URL ID is required", http.StatusBadRequest)
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	})
}
