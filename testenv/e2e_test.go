package testenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadE2EConfigFromEnv_Success(t *testing.T) {
	t.Setenv("SHIP_E2E_USER", "deploy")
	t.Setenv("SHIP_E2E_HOST", "staging.example.com")
	t.Setenv("SHIP_E2E_KEY", "/tmp/id_ed25519")

	cfg, err := LoadE2EConfigFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "deploy", cfg.User)
	assert.Equal(t, "staging.example.com", cfg.Host)
	assert.Equal(t, "/tmp/id_ed25519", cfg.KeyPath)
	assert.Equal(t, "deploy@staging.example.com", cfg.Address())
}

func TestLoadE2EConfigFromEnv_MissingVars(t *testing.T) {
	t.Setenv("SHIP_E2E_USER", "")
	t.Setenv("SHIP_E2E_HOST", "")
	t.Setenv("SHIP_E2E_KEY", "")

	_, err := LoadE2EConfigFromEnv()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SHIP_E2E_USER")
	assert.Contains(t, err.Error(), "SHIP_E2E_HOST")
	assert.Contains(t, err.Error(), "SHIP_E2E_KEY")
}
