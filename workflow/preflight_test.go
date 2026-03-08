package workflow

import (
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
	assert.Contains(t, err.Error(), "SSH key file not found: /tmp/no-such-key-file-abc123 — verify the --key path")
}

func TestCheckKeyFile_Exists(t *testing.T) {
	f, err := os.CreateTemp("", "ship-test-key")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	assert.NoError(t, checkKeyFile(f.Name()))
}

func TestCheckKeyFile_IncludesPath(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/tmp/no-such-key")
}

func TestCheckKeyFile_ReferencesFlag(t *testing.T) {
	err := checkKeyFile("/tmp/no-such-key")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--key")
}

func TestCheckKeyFile_DirectoryInsteadOfFile(t *testing.T) {
	dir := t.TempDir()

	err := checkKeyFile(dir)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "SSH key file not found")
	assert.Contains(t, err.Error(), "--key")
}

func TestCheckKeyFile_EmptyPath(t *testing.T) {
	err := checkKeyFile("")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "--key")
}

func TestCheckKeyFile_UnreadableFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("cannot test permission denial as root")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "unreadable-key")
	require.NoError(t, os.WriteFile(path, []byte("test"), 0o000))
	// On some systems, root can still read 0o000 files.
	// We check that if os.Stat fails with permission error, the message is correct.
	// os.Stat typically succeeds even for 0o000 files (only reads metadata), so
	// we test this by removing directory execute permission.
	require.NoError(t, os.Chmod(dir, 0o000))
	defer os.Chmod(dir, 0o755) //nolint:errcheck // cleanup

	err := checkKeyFile(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check file permissions")
}

func TestCheckComposeFiles_AllExist(t *testing.T) {
	f1, err := os.CreateTemp("", "ship-compose-1-*.yml")
	require.NoError(t, err)
	defer os.Remove(f1.Name())
	f1.Close()

	f2, err := os.CreateTemp("", "ship-compose-2-*.yml")
	require.NoError(t, err)
	defer os.Remove(f2.Name())
	f2.Close()

	assert.NoError(t, checkComposeFiles([]string{f1.Name(), f2.Name()}))
}

func TestCheckComposeFiles_FirstMissing(t *testing.T) {
	f, err := os.CreateTemp("", "ship-compose-*.yml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	err = checkComposeFiles([]string{"/tmp/nonexistent.yml", f.Name()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Compose file not found: /tmp/nonexistent.yml")
}

func TestCheckComposeFiles_SecondMissing(t *testing.T) {
	f, err := os.CreateTemp("", "ship-compose-*.yml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	err = checkComposeFiles([]string{f.Name(), "/tmp/nonexistent.yml"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "Compose file not found: /tmp/nonexistent.yml")
}

func TestCheckComposeFiles_SingleFile(t *testing.T) {
	f, err := os.CreateTemp("", "ship-compose-*.yml")
	require.NoError(t, err)
	defer os.Remove(f.Name())
	f.Close()

	assert.NoError(t, checkComposeFiles([]string{f.Name()}))
}

func TestCheckComposeFiles_ErrorIncludesPath(t *testing.T) {
	err := checkComposeFiles([]string{"/tmp/no-such-compose.yml"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "/tmp/no-such-compose.yml")
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
	// Check for standalone first-person words.
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
