package cli

import (
	"flag"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse_DefaultSSHStyleArgs(t *testing.T) {
	cfg, err := Parse([]string{"deploy@10.0.0.5", "app:latest"})

	require.NoError(t, err)
	assert.Equal(t, "deploy", cfg.User)
	assert.Equal(t, "10.0.0.5", cfg.Host)
	assert.Equal(t, []string{"app:latest"}, cfg.Images)
	assert.Equal(t, "", cfg.KeyPath)
	assert.Equal(t, 22, cfg.Port)
}

func TestParse_WithIdentityAndPort(t *testing.T) {
	cfg, err := Parse([]string{"-i", "~/.ssh/id_ed25519", "-p", "2222", "root@example.com", "ghcr.io/acme/app:dev", "redis:7"})

	require.NoError(t, err)
	assert.Equal(t, "root", cfg.User)
	assert.Equal(t, "example.com", cfg.Host)
	assert.Equal(t, []string{"ghcr.io/acme/app:dev", "redis:7"}, cfg.Images)
	assert.Equal(t, "~/.ssh/id_ed25519", cfg.KeyPath)
	assert.Equal(t, 2222, cfg.Port)
}

func TestParse_AllArgumentsMissing(t *testing.T) {
	_, err := Parse(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments: <user@host>, <image[:tag]>")
	assert.Contains(t, err.Error(), usageLine)
}

func TestParse_ImageMissing(t *testing.T) {
	_, err := Parse([]string{"deploy@example.com"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing required arguments: <image[:tag]>")
}

func TestParse_InvalidTarget(t *testing.T) {
	_, err := Parse([]string{"example.com", "app:latest"})

	require.Error(t, err)
	assert.Equal(t, "invalid target: example.com — expected <user@host>", err.Error())
}

func TestParse_MultipleImages(t *testing.T) {
	cfg, err := Parse([]string{"deploy@example.com", "app:latest", "traefik:v3", "redis:7"})

	require.NoError(t, err)
	assert.Equal(t, []string{"app:latest", "traefik:v3", "redis:7"}, cfg.Images)
}

func TestParse_EmptyIdentityFlag(t *testing.T) {
	_, err := Parse([]string{"-i", "", "deploy@example.com", "app:latest"})

	require.Error(t, err)
	assert.Equal(t, "empty -i flag — provide the path to an SSH private key", err.Error())
}

func TestParse_InvalidPort(t *testing.T) {
	_, err := Parse([]string{"-p", "0", "deploy@example.com", "app:latest"})

	require.Error(t, err)
	assert.Equal(t, "invalid -p value: 0 — port must be greater than 0", err.Error())
}

func TestParse_Help(t *testing.T) {
	_, err := Parse([]string{"--help"})

	require.ErrorIs(t, err, flag.ErrHelp)
}

func TestHelpText_MatchesDesignSpec(t *testing.T) {
	assert.Contains(t, HelpText, "ship — transfer local Docker images to a remote host over SSH")
	assert.Contains(t, HelpText, "Usage:")
	assert.Contains(t, HelpText, "<user@host> <image[:tag]> [<image[:tag]>...]")
	assert.Contains(t, HelpText, "-i, --identity-file <path>")
	assert.Contains(t, HelpText, "-p, --port <port>")
	assert.Contains(t, HelpText, "Examples:")
}

func TestUsageLine(t *testing.T) {
	_, err := Parse(nil)

	require.Error(t, err)
	lines := strings.SplitN(err.Error(), "\n", 2)
	require.Len(t, lines, 2)
	assert.Equal(t, usageLine, lines[1])
}
