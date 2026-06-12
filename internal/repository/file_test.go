package repository

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/Irongoshan-ux/url-shortener/internal/model"
	"github.com/stretchr/testify/require"
)

func TestFileRepositoryClosePersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "urls.json")

	repo, err := NewFileRepository(path)
	require.NoError(t, err)

	url := &model.URL{
		OriginalURL: "https://example.com",
		ShortURL:    "abc123",
		UserID:      "user-1",
	}
	require.NoError(t, repo.Create(context.Background(), url))
	require.NoError(t, repo.Close())

	reloaded, err := NewFileRepository(path)
	require.NoError(t, err)

	got, err := reloaded.GetByShortURL(context.Background(), "abc123")
	require.NoError(t, err)
	require.Equal(t, "https://example.com", got.OriginalURL)
	require.Equal(t, "user-1", got.UserID)
}
