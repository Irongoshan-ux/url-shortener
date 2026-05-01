package audit

import (
	"bytes"
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

const httpAuditTimeout = 10 * time.Second

type HTTPObserver struct {
	url    string
	client *http.Client
	log    zerolog.Logger
}

func NewHTTPObserver(endpoint string, log zerolog.Logger) *HTTPObserver {
	return &HTTPObserver{
		url: endpoint,
		client: &http.Client{
			Timeout: httpAuditTimeout,
		},
		log: log,
	}
}

func (h *HTTPObserver) OnAudit(ctx context.Context, e Event) {
	_ = ctx
	body, err := e.MarshalJSONLine()
	if err != nil {
		h.log.Error().Err(err).Msg("audit http: marshal event")
		return
	}
	reqCtx, cancel := context.WithTimeout(context.Background(), httpAuditTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, h.url, bytes.NewReader(body))
	if err != nil {
		h.log.Error().Err(err).Msg("audit http: new request")
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.client.Do(req)
	if err != nil {
		h.log.Error().Err(err).Str("url", h.url).Msg("audit http: request failed")
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		h.log.Error().Int("status", resp.StatusCode).Str("url", h.url).Msg("audit http: unexpected status")
	}
}
