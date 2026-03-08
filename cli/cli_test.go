package cli

import (
	"flag"
	"strings"
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
	assert.Equal(t, []string{"compose.yml"}, cfg.ComposeFiles)
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

func TestParse_MissingFlagsFormat(t *testing.T) {
	_, err := Parse([]string{"--user", "deploy", "--host", "example.com", "--key", "key.pem", "--command", "echo"})

	require.Error(t, err)
	msg := err.Error()
	// Single line for missing flags, not one per flag.
	assert.Contains(t, msg, "Missing required flags: --docker-compose")
}

func TestParse_MissingFlagsUsageLine(t *testing.T) {
	_, err := Parse([]string{})

	require.Error(t, err)
	msg := err.Error()
	lines := strings.SplitN(msg, "\n", 2)
	require.Len(t, lines, 2)
	assert.Equal(t, "Usage: ship --docker-compose <path> --user <user> --host <host> --key <path> --command <cmd>", lines[1])
}

func TestParse_Help(t *testing.T) {
	_, err := Parse([]string{"--help"})

	require.ErrorIs(t, err, flag.ErrHelp)
}

func TestParse_CommaSeparatedComposeFiles(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml,compose.prod.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"compose.yml", "compose.prod.yml"}, cfg.ComposeFiles)
}

func TestParse_SingleComposeFile(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"compose.yml"}, cfg.ComposeFiles)
}

func TestParse_ComposeFilesWhitespaceTrimmed(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml, compose.prod.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "docker compose up -d",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"compose.yml", "compose.prod.yml"}, cfg.ComposeFiles)
}

func TestParse_EmptyCommandMessage(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.Equal(t, "Empty --command flag \u2014 provide the command to run on the remote host", err.Error())
}

func TestParse_EmptyCommandExitsBeforeWorkflow(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "Missing required flags")
}

func TestParse_EmptyComposeFiles(t *testing.T) {
	args := []string{
		"--docker-compose", "",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "echo hi",
	}

	_, err := Parse(args)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Missing required flags: --docker-compose")
}

func TestParse_TrailingComma(t *testing.T) {
	args := []string{
		"--docker-compose", "compose.yml,",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "echo hi",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"compose.yml"}, cfg.ComposeFiles)
}

func TestParse_ThreeComposeFiles(t *testing.T) {
	args := []string{
		"--docker-compose", "a.yml,b.yml,c.yml",
		"--user", "deploy",
		"--host", "10.0.0.5",
		"--key", "~/.ssh/id_ed25519",
		"--command", "echo hi",
	}

	cfg, err := Parse(args)

	require.NoError(t, err)
	assert.Equal(t, []string{"a.yml", "b.yml", "c.yml"}, cfg.ComposeFiles)
}

func TestHelpText_MatchesDesignSpec(t *testing.T) {
	_, err := Parse([]string{"--help"})

	require.ErrorIs(t, err, flag.ErrHelp)

	expected := "ship \u2014 build, transfer, and deploy Docker Compose images to a remote host"
	assert.Contains(t, HelpText, expected)
	assert.Contains(t, HelpText, "Required Flags:")
	assert.Contains(t, HelpText, "Examples:")
}

func TestHelpText_ContainsMultiFileExample(t *testing.T) {
	assert.Contains(t, HelpText, "compose.yml,compose.prod.yml")
}

func TestHelpText_AllFlagsDocumented(t *testing.T) {
	assert.Contains(t, HelpText, "--docker-compose")
	assert.Contains(t, HelpText, "--user")
	assert.Contains(t, HelpText, "--host")
	assert.Contains(t, HelpText, "--key")
	assert.Contains(t, HelpText, "--command")
	assert.Contains(t, HelpText, "Path to compose file(s), comma-separated for multiple")
	assert.Contains(t, HelpText, "SSH user on the remote host")
	assert.Contains(t, HelpText, "Remote host address")
	assert.Contains(t, HelpText, "Path to SSH private key file")
	assert.Contains(t, HelpText, "Command to execute on the remote host after transfer")
}
