package audit

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

type FileObserver struct {
	path   string
	log    zerolog.Logger
	file   *os.File
	mu     sync.Mutex
	closed bool
}

func NewFileObserver(path string, log zerolog.Logger) (*FileObserver, error) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &FileObserver{path: path, log: log, file: f}, nil
}

func (f *FileObserver) OnAudit(_ context.Context, e Event) {
	line, err := json.Marshal(e)
	if err != nil {
		f.log.Error().Err(err).Msg("audit file: marshal event")
		return
	}
	line = append(line, '\n')

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return
	}
	if _, err := f.file.Write(line); err != nil {
		f.log.Error().Err(err).Str("path", f.path).Msg("audit file: write")
	}
}

// Close releases the audit file; safe to call more than once.
func (f *FileObserver) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.closed {
		return nil
	}
	f.closed = true
	return f.file.Close()
}
