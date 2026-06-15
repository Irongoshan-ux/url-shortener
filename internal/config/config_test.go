package config

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/stretchr/testify/require"
)

func testLoad(t *testing.T, args []string) (*Config, error) {
	t.Helper()
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	return load(fs, args)
}

func clearConfigEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"CONFIG", "SERVER_ADDRESS", "BASE_URL", "FILE_STORAGE_PATH",
		"DATABASE_DSN", "COOKIE_SECRET", "ENABLE_HTTPS", "AUDIT_FILE", "AUDIT_URL", "TRUSTED_SUBNET", "GRPC_SERVER",
	} {
		require.NoError(t, os.Unsetenv(key))
	}
}

func writeConfigFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

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

func TestLoadFromJSONFile(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{
		"server_address": "localhost:9090",
		"base_url": "http://localhost:9090",
		"file_storage_path": "/tmp/urls.json",
		"database_dsn": "postgres://localhost/db",
		"cookie_secret": "secret",
		"enable_https": true,
		"audit_file": "/tmp/audit.log",
		"audit_url": "http://audit.example/hook",
		"trusted_subnet": "192.168.1.0/24",
		"grpc_server": "localhost:50051"
	}`)

	cfg, err := testLoad(t, []string{"-c", path})
	require.NoError(t, err)
	require.Equal(t, "localhost:9090", cfg.ServerAddress)
	require.Equal(t, "http://localhost:9090", cfg.BaseURL)
	require.Equal(t, "/tmp/urls.json", cfg.FileStoragePath)
	require.Equal(t, "postgres://localhost/db", cfg.DatabaseDSN)
	require.Equal(t, "secret", cfg.CookieSecret)
	require.True(t, cfg.EnableHTTPS)
	require.Equal(t, "/tmp/audit.log", cfg.AuditFile)
	require.Equal(t, "http://audit.example/hook", cfg.AuditURL)
	require.Equal(t, "192.168.1.0/24", cfg.TrustedSubnet)
	require.Equal(t, "localhost:50051", cfg.GRPCServer)
}

func TestTrustedSubnetFromFlag(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := testLoad(t, []string{"-t", "10.0.0.0/8"})
	require.NoError(t, err)
	require.Equal(t, "10.0.0.0/8", cfg.TrustedSubnet)
}

func TestTrustedSubnetEnvOverridesFlag(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("TRUSTED_SUBNET", "172.16.0.0/12")

	cfg, err := testLoad(t, []string{"-t", "10.0.0.0/8"})
	require.NoError(t, err)
	require.Equal(t, "172.16.0.0/12", cfg.TrustedSubnet)
}

func TestGRPCServerFromFlag(t *testing.T) {
	clearConfigEnv(t)

	cfg, err := testLoad(t, []string{"-g", "localhost:50052"})
	require.NoError(t, err)
	require.Equal(t, "localhost:50052", cfg.GRPCServer)
}

func TestGRPCServerEnvOverridesFlag(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("GRPC_SERVER", "localhost:50053")

	cfg, err := testLoad(t, []string{"-g", "localhost:50052"})
	require.NoError(t, err)
	require.Equal(t, "localhost:50053", cfg.GRPCServer)
}

func TestLoadPartialJSONKeepsDefaults(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{"enable_https": true}`)

	cfg, err := testLoad(t, []string{"-config", path})
	require.NoError(t, err)
	require.Equal(t, "localhost:8080", cfg.ServerAddress)
	require.Equal(t, "http://localhost:8080", cfg.BaseURL)
	require.True(t, cfg.EnableHTTPS)
}

func TestFlagOverridesJSON(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{"server_address": "localhost:9090"}`)

	cfg, err := testLoad(t, []string{"-c", path, "-a", "localhost:8080"})
	require.NoError(t, err)
	require.Equal(t, "localhost:8080", cfg.ServerAddress)
}

func TestEnvOverridesJSONAndFlag(t *testing.T) {
	clearConfigEnv(t)
	t.Setenv("SERVER_ADDRESS", "localhost:7000")

	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{"server_address": "localhost:9090"}`)

	cfg, err := testLoad(t, []string{"-c", path, "-a", "localhost:8080"})
	require.NoError(t, err)
	require.Equal(t, "localhost:7000", cfg.ServerAddress)
}

func TestConfigPathFromEnv(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{"server_address": "localhost:9090"}`)
	t.Setenv("CONFIG", path)

	cfg, err := testLoad(t, nil)
	require.NoError(t, err)
	require.Equal(t, "localhost:9090", cfg.ServerAddress)
}

func TestConfigPathFlagOverridesEnv(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	envPath := writeConfigFile(t, dir, "env.json", `{"server_address": "localhost:1111"}`)
	flagPath := writeConfigFile(t, dir, "flag.json", `{"server_address": "localhost:2222"}`)
	t.Setenv("CONFIG", envPath)

	cfg, err := testLoad(t, []string{"-c", flagPath})
	require.NoError(t, err)
	require.Equal(t, "localhost:2222", cfg.ServerAddress)
}

func TestLoadMissingConfigFile(t *testing.T) {
	clearConfigEnv(t)

	_, err := testLoad(t, []string{"-c", "/no/such/config.json"})
	require.Error(t, err)
	require.Contains(t, err.Error(), "config file not found")
}

func TestLoadInvalidJSON(t *testing.T) {
	clearConfigEnv(t)
	dir := t.TempDir()
	path := writeConfigFile(t, dir, "config.json", `{invalid`)

	_, err := testLoad(t, []string{"-c", path})
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse config file")
}
