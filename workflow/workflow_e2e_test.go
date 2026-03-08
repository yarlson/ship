//go:build e2e

package workflow

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/progress"
	"ship/testlock"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func requireE2EPrereqs(t *testing.T, keyPath, user, host string) {
	t.Helper()

	dockerCmd := exec.CommandContext(context.Background(), "docker", "version")
	if err := dockerCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: Docker daemon unavailable: %v", err)
	}

	if _, err := os.Stat(keyPath); err != nil {
		t.Skipf("skipping e2e test: SSH key unavailable: %v", err)
	}

	sshCmd := exec.CommandContext(context.Background(), "ssh",
		"-i", keyPath,
		"-o", "ConnectTimeout=5",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		user+"@"+host,
		"true",
	)
	if err := sshCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: SSH test host unavailable: %v", err)
	}
}

func testSSHConfig(t *testing.T) (keyPath, user, host string) {
	t.Helper()
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	keyPath = home + "/.ssh/id_rsa"
	user = "root"
	host = "46.101.213.82"
	requireE2EPrereqs(t, keyPath, user, host)
	return keyPath, user, host
}

func setupComposeProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	dockerfile := filepath.Join(dir, "Dockerfile")
	require.NoError(t, os.WriteFile(dockerfile, []byte("FROM alpine:latest\nRUN echo hello\n"), 0o600))

	compose := filepath.Join(dir, "compose.yml")
	content := `services:
  web:
    build:
      context: .
      platforms:
        - linux/amd64
    image: ship-wftest-web:latest
    platform: linux/amd64
  api:
    build:
      context: .
      platforms:
        - linux/amd64
    image: ship-wftest-api:latest
    platform: linux/amd64
  redis:
    image: redis:alpine
`
	require.NoError(t, os.WriteFile(compose, []byte(content), 0o600))

	return compose
}

func TestPreflight_PassesWithValidConfig(t *testing.T) {
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo test",
	}

	err := Preflight(cfg)
	assert.NoError(t, err)
}

func TestRun_PrintsAllSevenStages(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// Progress output captured via progress.Writer contains only stage lines.
	for i := 1; i <= 7; i++ {
		pattern := regexp.MustCompile(`\[` + string(rune('0'+i)) + `/7\]`)
		assert.True(t, pattern.MatchString(out), "missing stage %d", i)
	}
}

func TestRun_StagesInOrder(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// Extract only stage progress lines from output.
	stagePattern := regexp.MustCompile(`^\[(\d)/7\]`)
	var stageLines []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if stagePattern.MatchString(line) {
			stageLines = append(stageLines, line)
		}
	}

	require.Len(t, stageLines, 14, "expected 14 stage lines (7 starts + 7 completes)")

	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5", "6", "6", "7", "7"}
	for i, line := range stageLines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2, "line %d did not match stage pattern: %s", i, line)
		assert.Equal(t, expectedOrder[i], matches[1], "line %d: expected stage %s, got %s", i, expectedOrder[i], matches[1])
	}
}

func TestRun_ReturnsNilOnSuccess(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployed",
	}

	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()

	err := Run(cfg)
	assert.NoError(t, err)
}

func TestWorkflow_FullSevenStages(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo deployment-complete",
	}

	out := captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// All 7 stage progress lines present.
	for i := 1; i <= 7; i++ {
		pattern := regexp.MustCompile(`\[` + string(rune('0'+i)) + `/7\]`)
		assert.True(t, pattern.MatchString(out), "missing stage %d in output", i)
	}

	// Success summary present.
	assert.Contains(t, out, "Ship complete")
	assert.Contains(t, out, "Host:     "+host)
	assert.Contains(t, out, "Command:  echo deployment-complete")
	assert.Contains(t, out, "Status:   Success")

	// Image names are original (not transfer tags).
	assert.Contains(t, out, "ship-wftest-web:latest")
	assert.Contains(t, out, "ship-wftest-api:latest")
	assert.NotContains(t, out, "localhost:5001/")
}

func TestWorkflow_TunnelCleanedUpOnSuccess(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "echo done",
	}

	captureOutput(func() {
		err := Run(cfg)
		require.NoError(t, err)
	})

	// After workflow completes, no tunnel processes should be running.
	out, err := exec.CommandContext(context.Background(), "bash", "-c", "ps aux | grep '[s]sh.*-R 5001' | grep -v grep || true").Output()
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(out)), "tunnel process should be cleaned up after success")
}

func TestWorkflow_TunnelCleanedUpOnFailure(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	composePath := setupComposeProject(t)
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		ComposeFiles: []string{composePath},
		User:         user,
		Host:         host,
		KeyPath:      keyPath,
		Command:      "exit 1",
	}

	captureOutput(func() {
		err := Run(cfg)
		assert.Error(t, err)
	})

	// After workflow error, no tunnel processes should be running.
	out, err := exec.CommandContext(context.Background(), "bash", "-c", "ps aux | grep '[s]sh.*-R 5001' | grep -v grep || true").Output()
	require.NoError(t, err)
	assert.Empty(t, strings.TrimSpace(string(out)), "tunnel process should be cleaned up after failure")
}
