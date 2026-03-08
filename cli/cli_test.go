package cli

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_AllFlagsPresent(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, "compose.yml", cfg.ComposeFiles)
	assert.Equal(t, "deploy", cfg.User)
	assert.Equal(t, "10.0.0.5", cfg.Host)
	assert.Equal(t, "~/.ssh/id_ed25519", cfg.KeyPath)
	assert.Equal(t, "docker compose up -d", cfg.Command)
}

func TestParse_AllFlagsMissing(t *testing.T) {
	_, err := Parse([]string{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required flags:")
	assert.Contains(t, err.Error(), "--docker-compose")
	assert.Contains(t, err.Error(), "--user")
	assert.Contains(t, err.Error(), "--host")
	assert.Contains(t, err.Error(), "--key")
	assert.Contains(t, err.Error(), "--command")
}

func TestParse_SomeFlagsMissing(t *testing.T) {
	args := []string{
		"--host", "10.0.0.5",
		"--user", "deploy",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--docker-compose")
	assert.Contains(t, err.Error(), "--key")
	assert.Contains(t, err.Error(), "--command")
	assert.NotContains(t, err.Error(), "--host,")
	assert.NotContains(t, err.Error(), "--user,")
}

func TestParse_EmptyCommand(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--command")
}

func TestParse_MissingFlagError_IncludesUsageLine(t *testing.T) {
	_, err := Parse([]string{"--host", "example.com"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Usage: ship")
}

func TestParse_OnlyOneFlag(t *testing.T) {
	_, err := Parse([]string{"--host", "example.com"})

	require.Error(t, err)
	msg := err.Error()
	// Exactly 4 missing flags.
	assert.Contains(t, msg, "--docker-compose")
	assert.Contains(t, msg, "--user")
	assert.Contains(t, msg, "--key")
	assert.Contains(t, msg, "--command")
}

func TestParse_WhitespaceOnlyCommand(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "   ",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--command")
}

func TestParse_Help(t *testing.T) {
	_, err := Parse([]string{"--help"})

	require.ErrorIs(t, err, flag.ErrHelp)
}
