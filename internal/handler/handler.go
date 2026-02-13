package handler

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
	"github.com/go-chi/chi/v5"
)

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Result string `json:"result"`
}

type Handler struct {
	service URLService
	baseURL string
}

func NewHandler(svc URLService, baseURL string) *Handler {
	return &Handler{
		service: svc,
		baseURL: baseURL,
	}
}

func (h *Handler) buildFullURL(r *http.Request, shortID string) (string, error) {
	if h.baseURL != "" {
		return url.JoinPath(h.baseURL, shortID)
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	base := scheme + "://" + r.Host
	return url.JoinPath(base, shortID)
}

func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
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

	parsedURL, err := validation.ParseHTTPURL(originalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	normalizedURL := parsedURL.String()

	shortURL, err := h.service.ShortenURL(r.Context(), normalizedURL)
	if err != nil {
		log.Printf("shorten url failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullURL, err := h.buildFullURL(r, shortURL.ShortURL)
	if err != nil {
		http.Error(w, "Failed to build short URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

func (h *Handler) ShortenURLJSON(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	originalURL := strings.TrimSpace(req.URL)
	if originalURL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	parsedURL, err := validation.ParseHTTPURL(originalURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	normalizedURL := parsedURL.String()

	shortURL, err := h.service.ShortenURL(r.Context(), normalizedURL)
	if err != nil {
		log.Printf("shorten url failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	fullURL, err := h.buildFullURL(r, shortURL.ShortURL)
	if err != nil {
		http.Error(w, "Failed to build short URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(shortenResponse{Result: fullURL})
}

type batchRequestItem struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type batchResponseItem struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *Handler) ShortenURLBatch(w http.ResponseWriter, r *http.Request) {
	var req []batchRequestItem
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if len(req) == 0 {
		http.Error(w, "empty batch", http.StatusBadRequest)
		return
	}

	items := make([]service.BatchItem, 0, len(req))
	for _, x := range req {
		originalURL := strings.TrimSpace(x.OriginalURL)
		if originalURL == "" {
			http.Error(w, "original_url is required", http.StatusBadRequest)
			return
		}
		parsedURL, err := validation.ParseHTTPURL(originalURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items = append(items, service.BatchItem{CorrelationID: x.CorrelationID, OriginalURL: parsedURL.String()})
	}

	results, err := h.service.ShortenURLBatch(r.Context(), items)
	if err != nil {
		log.Printf("shorten url batch failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	resp := make([]batchResponseItem, len(results))
	for i, res := range results {
		fullURL, err := h.buildFullURL(r, res.ShortURL)
		if err != nil {
			http.Error(w, "Failed to build short URL", http.StatusInternalServerError)
			return
		}
		resp[i] = batchResponseItem{CorrelationID: res.CorrelationID, ShortURL: fullURL}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Short URL ID is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.service.GetOriginalURL(r.Context(), shortID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			http.Error(w, "URL not found", http.StatusBadRequest)
			return
		}

		log.Printf("get original url failed: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJSON)
	r.Post("/api/shorten/batch", h.ShortenURLBatch)

	r.Get("/{id}", h.Redirect)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Short URL ID is required", http.StatusBadRequest)
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad request", http.StatusBadRequest)
	})

	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			http.Error(w, "Short URL ID is required", http.StatusBadRequest)
		} else {
			http.Error(w, "Bad request", http.StatusBadRequest)
		}
	})

	return r
}
