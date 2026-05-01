package audit

import (
	"context"
	"os"
	"sync"

	"github.com/rs/zerolog"
)

type FileObserver struct {
	path string
	log  zerolog.Logger
	mu   sync.Mutex
}

func NewFileObserver(path string, log zerolog.Logger) *FileObserver {
	return &FileObserver{path: path, log: log}
}

func (f *FileObserver) OnAudit(ctx context.Context, e Event) {
	_ = ctx
	f.mu.Lock()
	defer f.mu.Unlock()
	line, err := e.MarshalJSONLine()
	if err != nil {
		f.log.Error().Err(err).Msg("audit file: marshal event")
		return
	}
	file, err := os.OpenFile(f.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		f.log.Error().Err(err).Str("path", f.path).Msg("audit file: open")
		return
	}
	defer file.Close()
	if _, err := file.Write(append(line, '\n')); err != nil {
		f.log.Error().Err(err).Str("path", f.path).Msg("audit file: write")
	}
}
