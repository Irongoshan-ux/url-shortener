// Package handler implements HTTP handlers for the URL shortener REST API (plain-text and JSON shorten, batch, redirect, user URLs).
package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/Irongoshan-ux/url-shortener/internal/audit"
	"github.com/Irongoshan-ux/url-shortener/internal/auth"
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

type statsResponse struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

// Handler serves HTTP requests using ShortenerFacade. baseURL is configured on the facade.
type Handler struct {
	facade        *ShortenerFacade
	log           zerolog.Logger
	audit         *audit.Subject
	trustedSubnet string
}

// NewHandler constructs a Handler. auditSubject may be nil; log is used for server-side errors and diagnostics.
func NewHandler(facade *ShortenerFacade, log zerolog.Logger, auditSubject *audit.Subject, trustedSubnet string) *Handler {
	return &Handler{
		facade:        facade,
		log:           log,
		audit:         auditSubject,
		trustedSubnet: trustedSubnet,
	}
}

func (h *Handler) notifyAudit(action, originalURL string, r *http.Request) {
	if h.audit == nil {
		return
	}
	var userID string
	if uid, err := auth.UserIDFromContext(r.Context()); err == nil && uid != "" {
		userID = uid
	}
	h.audit.Notify(context.Background(), audit.NewEvent(action, userID, originalURL))
}

func (h *Handler) buildFullURL(r *http.Request, shortID string) (string, error) {
	return h.facade.buildFullURL(r, shortID)
}

// ShortenURL handles POST / with a plain-text body containing the original URL. Requires a user id in context (via auth middleware). On success responds with 201 and the short URL as text/plain. Returns 409 with the existing short URL if the original URL was already shortened for this user.
func (h *Handler) ShortenURL(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	originalURL := string(bytes.TrimSpace(body))
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
	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("user id not in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortID, conflict, err := h.facade.ShortenURL(r.Context(), normalizedURL, userID)
	if err != nil {
		h.log.Info().Err(err).Msg("shorten url failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fullURL, err := h.buildFullURL(r, shortID)
	if err != nil {
		http.Error(w, "Failed to build short URL", http.StatusInternalServerError)
		return
	}
	if conflict {
		h.notifyAudit(audit.ActionShorten, normalizedURL, r)
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(fullURL))
		return
	}

	h.notifyAudit(audit.ActionShorten, normalizedURL, r)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fullURL))
}

// ShortenURLJSON handles POST /api/shorten with JSON body {"url":"..."}. Response is JSON {"result":"<short url>"} with 201 on success, or 409 with the same shape if the URL already exists.
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

	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("user id not in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	shortID, conflict, err := h.facade.ShortenURL(r.Context(), originalURL, userID)
	if err != nil {
		if errors.Is(err, ErrEmptyURL) || errors.Is(err, ErrInvalidURL) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		h.log.Info().Err(err).Msg("shorten url failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fullURL, err := h.buildFullURL(r, shortID)
	if err != nil {
		http.Error(w, "Failed to build short URL", http.StatusInternalServerError)
		return
	}
	if conflict {
		h.notifyAudit(audit.ActionShorten, originalURL, r)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(shortenResponse{Result: fullURL})
		return
	}

	h.notifyAudit(audit.ActionShorten, originalURL, r)

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

// ShortenURLBatch handles POST /api/shorten/batch with a JSON array of {correlation_id, original_url}. Responds with 201 and an array of {correlation_id, short_url} with absolute short links.
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

	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("user id not in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	results, err := h.facade.service.ShortenURLBatch(r.Context(), items, userID)
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

// Redirect handles GET /{id}: looks up the short id and responds with 307 Temporary Redirect to the original URL, 404 if missing, or 410 if soft-deleted.
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	shortID := chi.URLParam(r, "id")
	if shortID == "" {
		http.Error(w, "Short URL ID is required", http.StatusBadRequest)
		return
	}

	originalURL, err := h.facade.ExpandURL(r.Context(), shortID)
	if err != nil {
		switch {
		case errors.Is(err, ErrExpandNotFound):
			w.WriteHeader(http.StatusNotFound)
			return
		case errors.Is(err, ErrExpandGone):
			w.WriteHeader(http.StatusGone)
			return
		default:
			h.log.Info().Err(err).Msg("get original url failed")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}

	h.notifyAudit(audit.ActionFollow, originalURL, r)

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}

type userURLItem struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// GetUserURLs handles GET /api/user/urls for the current user (valid cookie). Returns 200 and JSON array of {short_url, original_url}, 204 if there are no URLs or no valid cookie yet, or 401 if a cookie was sent but invalid.
func (h *Handler) GetUserURLs(w http.ResponseWriter, r *http.Request) {
	if !auth.HadValidCookieFromContext(r.Context()) {
		if auth.HadCookieInRequestFromContext(r.Context()) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}
	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil || userID == "" {
		h.log.Error().Err(err).Msg("user id not in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	urls, err := h.facade.ListUserURLs(r.Context(), userID, r)
	if err != nil {
		h.log.Info().Err(err).Msg("get user urls failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	if len(urls) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(urls)
}

// DeleteUserURLs handles DELETE /api/user/urls with a JSON array of short ids to soft-delete for the current user. Responds with 202 Accepted after enqueueing work; 401 without a valid cookie.
func (h *Handler) DeleteUserURLs(w http.ResponseWriter, r *http.Request) {
	if !auth.HadValidCookieFromContext(r.Context()) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	userID, err := auth.UserIDFromContext(r.Context())
	if err != nil || userID == "" {
		h.log.Error().Err(err).Msg("user id not in context")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	var shortIDs []string
	if err := json.NewDecoder(r.Body).Decode(&shortIDs); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	h.facade.service.DeleteUserURLs(r.Context(), userID, shortIDs)
	w.WriteHeader(http.StatusAccepted)
}

// GetStats handles GET /api/internal/stats for clients in the configured trusted subnet.
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	if !trustedSubnetAllowed(h.trustedSubnet, r) {
		http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		return
	}
	urls, users, err := h.facade.GetStats(r.Context())
	if err != nil {
		h.log.Error().Err(err).Msg("get stats failed")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(statsResponse{URLs: urls, Users: users}); err != nil {
		h.log.Error().Err(err).Msg("encode stats response failed")
	}
}

// Router registers all API routes on a new chi.Router: POST /, /api/shorten, /api/shorten/batch, GET/DELETE /api/user/urls, GET /{id}, plus root GET and not-found handlers.
func (h *Handler) Router() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.ShortenURL)
	r.Post("/api/shorten", h.ShortenURLJSON)
	r.Post("/api/shorten/batch", h.ShortenURLBatch)
	r.Get("/api/user/urls", h.GetUserURLs)
	r.Delete("/api/user/urls", h.DeleteUserURLs)
	r.Get("/api/internal/stats", h.GetStats)

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
