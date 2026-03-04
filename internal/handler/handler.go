package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/auth"
	"github.com/Irongoshan-ux/url-shortener/internal/repository"
	"github.com/Irongoshan-ux/url-shortener/internal/service"
	"github.com/Irongoshan-ux/url-shortener/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
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
	log     zerolog.Logger
}

func NewHandler(svc URLService, baseURL string, log zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		baseURL: baseURL,
		log:     log,
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
	userID, _ := auth.UserIDFromContext(r.Context())

	shortURL, err := h.service.ShortenURL(r.Context(), normalizedURL, userID)
	if err != nil {
		if errors.Is(err, service.ErrConflict) && shortURL != nil {
			fullURL, _ := h.buildFullURL(r, shortURL.ShortURL)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(fullURL))
			return
		}
		h.log.Info().Err(err).Msg("shorten url failed")
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
	userID, _ := auth.UserIDFromContext(r.Context())

	shortURL, err := h.service.ShortenURL(r.Context(), normalizedURL, userID)
	if err != nil {
		if errors.Is(err, service.ErrConflict) && shortURL != nil {
			fullURL, _ := h.buildFullURL(r, shortURL.ShortURL)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(shortenResponse{Result: fullURL})
			return
		}
		h.log.Info().Err(err).Msg("shorten url failed")
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

	userID, _ := auth.UserIDFromContext(r.Context())
	results, err := h.service.ShortenURLBatch(r.Context(), items, userID)
	if err != nil {
		h.log.Info().Err(err).Msg("shorten url batch failed")
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

	url, err := h.service.GetURLByShortID(r.Context(), shortID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		h.log.Info().Err(err).Msg("get original url failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if url.IsDeleted {
		w.WriteHeader(http.StatusGone)
		return
	}

	http.Redirect(w, r, url.OriginalURL, http.StatusTemporaryRedirect)
}

type userURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	if !auth.HadValidCookieFromContext(r.Context()) {
		if auth.HadCookieInRequestFromContext(r.Context()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	urls, err := h.service.GetUserURLs(r.Context(), userID)
	if err != nil {
		h.log.Info().Err(err).Msg("get user urls failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	resp := make([]userURLItem, len(urls))
	for i, u := range urls {
		fullShort, _ := h.buildFullURL(r, u.ShortURL)
		resp[i] = userURLItem{ShortURL: fullShort, OriginalURL: u.OriginalURL}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	if !auth.HadValidCookieFromContext(r.Context()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	h.service.DeleteUserURLs(r.Context(), userID, shortIDs)
	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJSON)
	r.Post("/api/shorten/batch", h.ShortenURLBatch)
	r.Get("/api/user/urls", h.GetUserURLs)
	r.Delete("/api/user/urls", h.DeleteUserURLs)

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
