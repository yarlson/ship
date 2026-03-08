package workflow

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ship/cli"
	"ship/progress"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func validConfig() cli.Config {
	return cli.Config{
		ComposeFiles: "compose.yml",
		User:         "deploy",
		Host:         "10.0.0.5",
		KeyPath:      "~/.ssh/id_ed25519",
		Command:      "docker compose up -d",
	}
}

func TestRun_PrintsAllSevenStages(t *testing.T) {
	out := captureOutput(func() {
		err := Run(validConfig())
		require.NoError(t, err)
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	assert.Len(t, lines, 14, "expected 14 lines (7 starts + 7 completes)")

	for i := 1; i <= 7; i++ {
		pattern := regexp.MustCompile(`\[` + string(rune('0'+i)) + `/7\]`)
		assert.True(t, pattern.MatchString(out), "missing stage %d", i)
	}
}

func TestRun_StagesInOrder(t *testing.T) {
	out := captureOutput(func() {
		err := Run(validConfig())
		require.NoError(t, err)
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	require.Len(t, lines, 14)

	stagePattern := regexp.MustCompile(`^\[(\d)/7\]`)
	expectedOrder := []string{"1", "1", "2", "2", "3", "3", "4", "4", "5", "5", "6", "6", "7", "7"}

	for i, line := range lines {
		matches := stagePattern.FindStringSubmatch(line)
		require.Len(t, matches, 2, "line %d did not match stage pattern: %s", i, line)
		assert.Equal(t, expectedOrder[i], matches[1], "line %d: expected stage %s, got %s", i, expectedOrder[i], matches[1])
	}
}

func TestRun_ReturnsNilOnSuccess(t *testing.T) {
	// Suppress output.
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()

	err := Run(validConfig())
	assert.NoError(t, err)
}
