package app

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

const (
	encodingGzip          = "gzip"
	contentEncodingHeader = "Content-Encoding"
	acceptEncodingHeader  = "Accept-Encoding"
	contentTypeHeader     = "Content-Type"
)

var compressibleContentTypes = map[string]bool{
	"application/json": true,
	"text/html":        true,
}

func isCompressible(contentType string) bool {
	base := contentType
	if i := strings.IndexByte(contentType, ';'); i >= 0 {
		base = strings.TrimSpace(contentType[:i])
	}
	return compressibleContentTypes[base]
}

func clientAcceptsGzip(ae string) bool {
	return strings.Contains(ae, encodingGzip)
}

type gzipResponseWriter struct {
	http.ResponseWriter
	request     *http.Request
	writer      io.Writer
	gzipWriter  *gzip.Writer
	wroteHeader bool
}

func (w *gzipResponseWriter) WriteHeader(code int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	contentType := w.Header().Get(contentTypeHeader)
	if isCompressible(contentType) && clientAcceptsGzip(w.request.Header.Get(acceptEncodingHeader)) {
		w.Header().Set(contentEncodingHeader, encodingGzip)
		w.Header().Del("Content-Length")
		w.gzipWriter = gzip.NewWriter(w.ResponseWriter)
		w.writer = w.gzipWriter
	} else {
		w.writer = w.ResponseWriter
	}

	w.ResponseWriter.WriteHeader(code)
}

func (w *gzipResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.writer.Write(p)
	return n, err
}

func (w *gzipResponseWriter) Close() error {
	if w.gzipWriter != nil {
		return w.gzipWriter.Close()
	}
	return nil
}

// GzipMiddleware transparently decompresses gzip request bodies and compresses eligible JSON/text responses when Accept-Encoding allows it.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(contentEncodingHeader) == encodingGzip {
			gr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			defer gr.Close()
			r.Body = gr
		}

		gw := &gzipResponseWriter{
			ResponseWriter: w,
			request:        r,
			writer:         w,
		}
		defer gw.Close()

		next.ServeHTTP(gw, r)
	})
}
