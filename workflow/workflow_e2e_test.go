//go:build e2e

package workflow

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/docker"
	"ship/progress"
	"ship/testctx"
	"ship/testenv"
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

func testSSHConfig(t *testing.T) (keyPath, user, host string) {
	t.Helper()
	cfg := testenv.RequireE2EConfig(t)
	return cfg.KeyPath, cfg.User, cfg.Host
}

func setupLocalImage(t *testing.T, imageRef string) {
	t.Helper()
	ctx := testctx.New(t)

	version := exec.CommandContext(ctx, "docker", "version")
	if err := version.Run(); err != nil {
		t.Skipf("skipping e2e test: Docker daemon unavailable: %v", err)
	}

	pull := exec.CommandContext(ctx, "docker", "pull", "alpine:latest")
	if out, err := pull.CombinedOutput(); err != nil {
		t.Skipf("skipping e2e test: failed to pull alpine: %v\n%s", err, string(out))
	}

	require.NoError(t, docker.TagImage(ctx, "alpine:latest", imageRef))
}

func TestPreflight_PassesWithValidConfig(t *testing.T) {
	imageRefs := []string{"ship-wftest-preflight:latest", "ship-wftest-preflight-proxy:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		Images:  imageRefs,
		User:    user,
		Host:    host,
		KeyPath: keyPath,
		Port:    22,
	}

	err := Preflight(testctx.New(t), cfg)
	assert.NoError(t, err)
}

func TestRun_PrintsAllFiveStages(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	imageRefs := []string{"ship-wftest-run:latest", "ship-wftest-run-proxy:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		Images:  imageRefs,
		User:    user,
		Host:    host,
		KeyPath: keyPath,
		Port:    22,
	}

	out := captureOutput(func() {
		err := Run(testctx.New(t), cfg)
		require.NoError(t, err)
	})

	for i := 1; i <= 5; i++ {
		pattern := regexp.MustCompile(`\[` + string(rune('0'+i)) + `/5\]`)
		assert.True(t, pattern.MatchString(out), "missing stage %d", i)
	}
}

func TestRun_StagesInOrder(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	imageRefs := []string{"ship-wftest-order:latest", "ship-wftest-order-proxy:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		Images:  imageRefs,
		User:    user,
		Host:    host,
		KeyPath: keyPath,
		Port:    22,
	}

	out := captureOutput(func() {
		err := Run(testctx.New(t), cfg)
		require.NoError(t, err)
	})

	stagePattern := regexp.MustCompile(`^\[(\d)/5\]`)
	var stageLines []string
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if stagePattern.MatchString(line) {
			stageLines = append(stageLines, line)
		}
	}

	require.Len(t, stageLines, 10, "expected 10 stage lines (5 starts + 5 completes)")

	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5"}
	for i, line := range stageLines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2)
		assert.Equal(t, expectedOrder[i], matches[1], "line %d", i)
	}
}

func TestWorkflow_Summary(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	imageRefs := []string{"ship-wftest-summary:latest", "ship-wftest-summary-proxy:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		Images:  imageRefs,
		User:    user,
		Host:    host,
		KeyPath: keyPath,
		Port:    22,
	}

	out := captureOutput(func() {
		err := Run(testctx.New(t), cfg)
		require.NoError(t, err)
	})

	assert.Contains(t, out, "Ship complete")
	assert.Contains(t, out, "Host:     "+host)
	assert.Contains(t, out, "Images:   "+strings.Join(imageRefs, ", "))
	assert.Contains(t, out, "Status:   Success")
	assert.NotContains(t, out, "localhost:5001/")
}

func TestWorkflow_TunnelCleanedUpOnSuccess(t *testing.T) {
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	imageRefs := []string{"ship-wftest-cleanup:latest", "ship-wftest-cleanup-proxy:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}
	keyPath, user, host := testSSHConfig(t)
	cfg := cli.Config{
		Images:  imageRefs,
		User:    user,
		Host:    host,
		KeyPath: keyPath,
		Port:    22,
	}

	err := Run(testctx.New(t), cfg)
	require.NoError(t, err)

	out, cmdErr := exec.CommandContext(testctx.New(t), "bash", "-c", "ps aux | grep '[s]sh.*-R 5001' | grep -v grep || true").Output()
	require.NoError(t, cmdErr)
	assert.Empty(t, strings.TrimSpace(string(out)), "tunnel process should be cleaned up after success")
}
