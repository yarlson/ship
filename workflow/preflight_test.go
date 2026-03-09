package workflow

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckKeyFile_NotFound(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key-file-abc123")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH key file not found: /tmp/no-such-key-file-abc123 — verify the -i path")
}

func TestCheckKeyFile_Exists(t *testing.T) {
	f, err := os.CreateTemp("", "ship-test-key")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	assert.NoError(t, checkKeyFile(f.Name()))
}

func TestCheckKeyFile_EmptyPathAllowed(t *testing.T) {
	assert.NoError(t, checkKeyFile(""))
}

func TestCheckKeyFile_DirectoryInsteadOfFile(t *testing.T) {
	dir := t.TempDir()

	err := checkKeyFile(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH key file not found")
	assert.Contains(t, err.Error(), "-i")
}

func TestCheckKeyFile_UnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test permission denial as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable-key")
	require.NoError(t, os.WriteFile(path, []byte("test"), 0o000))
	require.NoError(t, os.Chmod(dir, 0o000))
	defer os.Chmod(dir, 0o755) //nolint:errcheck // cleanup

	err := checkKeyFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check file permissions")
}

func TestPreflight_ErrorFormat_NoEmoji(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")
	require.Error(t, err)

	emoji := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	assert.False(t, emoji.MatchString(err.Error()))
}

func TestPreflight_ErrorFormat_NoANSI(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")
	require.Error(t, err)

	assert.NotContains(t, err.Error(), "\x1b[")
}

func TestPreflight_ErrorFormat_NoFirstPerson(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")
	require.Error(t, err)

	msg := err.Error()
	assert.NotRegexp(t, `\bI\b`, msg)
	assert.NotContains(t, msg, " we ")
	assert.NotContains(t, msg, " my ")
	assert.NotContains(t, msg, " our ")
}

func TestPreflight_ErrorFormat_NoHedging(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")
	require.Error(t, err)

	msg := err.Error()
	assert.NotContains(t, msg, "might")
	assert.NotContains(t, msg, "possibly")
	assert.NotContains(t, msg, "try to")
}

func TestPreflightCommandError_PreservesContextCanceled(t *testing.T) {
	err := preflightCommandError(context.Canceled, "docker is not installed or not in PATH")

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestPreflightCommandError_PreservesDeadlineExceeded(t *testing.T) {
	err := preflightCommandError(context.DeadlineExceeded, "SSH connection failed — verify the target and SSH credentials")

	require.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestPreflightCommandError_UsesFallbackForOtherFailures(t *testing.T) {
	err := preflightCommandError(errors.New("boom"), "docker is not installed or not in PATH")

	require.Error(t, err)
	assert.EqualError(t, err, "docker is not installed or not in PATH")
}
