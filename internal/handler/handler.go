package handler

import (
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/service"
)

type Handler struct {
	service *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		service: svc,
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

	// Validate that the URL is a valid absolute URL
	parsedURL, err := url.Parse(originalURL)
	if err != nil {
		http.Error(w, "Invalid URL format", http.StatusBadRequest)
		return
	}

	// Ensure it's an absolute URL with a scheme
	if parsedURL.Scheme == "" {
		http.Error(w, "URL must be absolute (include http:// or https://)", http.StatusBadRequest)
		return
	}

	// Ensure it has a host
	if parsedURL.Host == "" {
		http.Error(w, "URL must include a host", http.StatusBadRequest)
		return
	}

	// Only allow http and https schemes
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		http.Error(w, "URL scheme must be http or https", http.StatusBadRequest)
		return
	}

	shortURL, err := h.service.ShortenURL(r.Context(), originalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Construct the full URL from the request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	fullURL := scheme + "://" + r.Host + "/" + shortURL.ShortURL

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	originalURL, err := h.service.GetOriginalURL(r.Context(), path)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) Root(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")

	switch r.Method {
	case http.MethodGet:
		// GET / - return error, GET /{id} - redirect
		if path == "" {
			http.Error(w, "Short URL ID is required", http.StatusBadRequest)
			return
		}
		h.Redirect(w, r)
	case http.MethodPost:
		// POST / - shorten URL
		if path != "" {
			http.Error(w, "POST request should be to root path", http.StatusBadRequest)
			return
		}
		h.ShortenURL(w, r)
	default:
		http.Error(w, "Bad request", http.StatusBadRequest)
	}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", h.Root)
}
