package config

import (
	"os"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/stretchr/testify/require"
)

func TestEnableHTTPSFromEnv(t *testing.T) {
	t.Setenv("ENABLE_HTTPS", "true")

	var cfg Config
	require.NoError(t, cleanenv.ReadEnv(&cfg))
	require.True(t, cfg.EnableHTTPS)
}

func TestEnableHTTPSFromEnvFalse(t *testing.T) {
	t.Setenv("ENABLE_HTTPS", "false")

	var cfg Config
	require.NoError(t, cleanenv.ReadEnv(&cfg))
	require.False(t, cfg.EnableHTTPS)
}

func TestEnableHTTPSUnset(t *testing.T) {
	require.NoError(t, os.Unsetenv("ENABLE_HTTPS"))

	var cfg Config
	require.NoError(t, cleanenv.ReadEnv(&cfg))
	require.False(t, cfg.EnableHTTPS)
}
