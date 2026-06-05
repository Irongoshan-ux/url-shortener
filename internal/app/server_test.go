package app

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Irongoshan-ux/url-shortener/internal/config"
	"github.com/stretchr/testify/require"
)

func TestServeHTTPS(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	cfg := &config.Config{EnableHTTPS: true}

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(cfg, srv)
	}()

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // test-only self-signed cert
		},
		Timeout: 2 * time.Second,
	}

	var resp *http.Response
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = client.Get("https://" + addr + "/ping")
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())

	require.NoError(t, srv.Close())
	require.ErrorIs(t, <-errCh, http.ErrServerClosed)
}

func TestServeHTTP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	addr := ln.Addr().String()
	require.NoError(t, ln.Close())

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	cfg := &config.Config{EnableHTTPS: false}

	errCh := make(chan error, 1)
	go func() {
		errCh <- Serve(cfg, srv)
	}()

	client := &http.Client{Timeout: 2 * time.Second}

	var resp *http.Response
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		resp, err = client.Get("http://" + addr + "/ping")
		if err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	_, _ = io.Copy(io.Discard, resp.Body)
	require.NoError(t, resp.Body.Close())

	require.NoError(t, srv.Close())
	require.ErrorIs(t, <-errCh, http.ErrServerClosed)
}
