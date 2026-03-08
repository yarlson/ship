//go:build e2e

package main_test

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/docker"
	"ship/testenv"
	"ship/testlock"
)

func setupLocalImage(t *testing.T, imageRef string) {
	t.Helper()

	pull := exec.CommandContext(context.Background(), "docker", "pull", "alpine:latest")
	if out, err := pull.CombinedOutput(); err != nil {
		t.Fatalf("failed to pull alpine: %v\n%s", err, string(out))
	}

	require.NoError(t, docker.TagImage("alpine:latest", imageRef))
}

func TestShip_PrintsFiveStages(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)
	testlock.Port5001(t)
	testlock.StopRegistry(t)
	t.Cleanup(func() { testlock.StopRegistry(t) })
	imageRefs := []string{"ship-main-e2e:latest", "traefik:v3"}
	for _, imageRef := range imageRefs {
		setupLocalImage(t, imageRef)
	}

	args := []string{}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	args = append(args, cfg.Address())
	args = append(args, imageRefs...)

	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	out, err := cmd.Output()
	require.NoError(t, err)

	stdout := string(out)
	lines := strings.Split(strings.TrimSpace(stdout), "\n")

	var stageLines []string
	stagePattern := regexp.MustCompile(`^\[(\d)/5\]`)
	for _, line := range lines {
		if stagePattern.MatchString(line) {
			stageLines = append(stageLines, line)
		}
	}

	assert.Len(t, stageLines, 10, "expected 10 stage lines (5 starts + 5 completes)")

	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5"}
	for i, line := range stageLines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2)
		assert.Equal(t, expectedOrder[i], matches[1])
	}

	assert.Contains(t, stdout, "Ship complete")
	assert.Contains(t, stdout, "Host:     "+cfg.Host)
	assert.Contains(t, stdout, "Images:   "+strings.Join(imageRefs, ", "))
	assert.NotContains(t, stdout, "localhost:5001/")
}

func TestShip_BadKeyPath_FailsBeforeStages(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)
	imageRef := "ship-main-bad-key:latest"
	setupLocalImage(t, imageRef)

	cmd := exec.CommandContext(context.Background(), binaryPath, "-i", "/tmp/nonexistent-key-ship-test", cfg.Address(), imageRef)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err)
	errOut := stderr.String()
	assert.Contains(t, errOut, "SSH key file not found")
	assert.Contains(t, errOut, "/tmp/nonexistent-key-ship-test")
	assert.Contains(t, errOut, "-i")
	assert.NotContains(t, stdout.String(), "[1/5]")
}

func TestShip_UnreachableHost_FailsBeforeStages(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)
	imageRef := "ship-main-unreachable:latest"
	setupLocalImage(t, imageRef)

	args := []string{}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	args = append(args, "root@192.0.2.1", imageRef)

	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err)
	assert.Contains(t, stderr.String(), "SSH connection failed")
	assert.NotContains(t, stdout.String(), "[1/5]")
}

func TestShip_MissingLocalImage_FailsBeforeStages(t *testing.T) {
	cfg := testenv.RequireE2EConfig(t)
	args := []string{}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	args = append(args, cfg.Address(), "ship-main-missing:latest")

	cmd := exec.CommandContext(context.Background(), binaryPath, args...)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	require.Error(t, err)
	assert.Contains(t, stderr.String(), "local image not found: ship-main-missing:latest")
	assert.NotContains(t, stdout.String(), "[1/5]")
}
