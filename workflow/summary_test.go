package workflow

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"ship/cli"
	"ship/progress"
)

func captureSummaryOutput(fn func()) string {
	var buf bytes.Buffer
	orig := progress.Writer
	progress.Writer = &buf
	defer func() { progress.Writer = orig }()
	fn()
	return buf.String()
}

func TestPrintSummary_Format(t *testing.T) {
	cfg := cli.Config{
		Host:    "h.example.com",
		Command: "docker compose up -d",
	}
	state := &State{
		ImageMap: map[string]string{
			"app:latest":    "localhost:5001/app:latest",
			"worker:latest": "localhost:5001/worker:latest",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "Ship complete")
	assert.Contains(t, out, "Host:     h.example.com")
	assert.Contains(t, out, "Command:  docker compose up -d")
	assert.Contains(t, out, "Status:   Success")

	// Both image names should appear (order may vary with map iteration).
	assert.Contains(t, out, "app:latest")
	assert.Contains(t, out, "worker:latest")
}

func TestPrintSummary_SingleImage(t *testing.T) {
	cfg := cli.Config{
		Host:    "host",
		Command: "cmd",
	}
	state := &State{
		ImageMap: map[string]string{
			"web:latest": "localhost:5001/web:latest",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "Images:   web:latest")
	// No trailing comma.
	assert.NotContains(t, out, "web:latest,")
}

func TestPrintSummary_LabelAlignment(t *testing.T) {
	cfg := cli.Config{
		Host:    "host",
		Command: "cmd",
	}
	state := &State{
		ImageMap: map[string]string{
			"a:1": "localhost:5001/a:1",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Lines 1-4 are the label lines (0 is "Ship complete").
	for _, line := range lines[1:] {
		// All value columns should start at the same position.
		// "  Host:     " is 12 chars, so values start at column 12.
		assert.True(t, len(line) >= 12, "line too short: %s", line)
		prefix := line[:12]
		assert.True(t, strings.HasSuffix(strings.TrimRight(prefix, " "), ":"),
			"label should end with colon: %q", prefix)
	}
}

func TestPrintSummary_NoTransferTagsInOutput(t *testing.T) {
	cfg := cli.Config{
		Host:    "host",
		Command: "cmd",
	}
	state := &State{
		ImageMap: map[string]string{
			"app:latest": "localhost:5001/app:latest",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "app:latest")
	assert.NotContains(t, out, "localhost:5001/")
}

func TestPrintSummary_ManyImages(t *testing.T) {
	cfg := cli.Config{
		Host:    "host",
		Command: "cmd",
	}
	state := &State{
		ImageMap: map[string]string{
			"a:1": "localhost:5001/a:1",
			"b:2": "localhost:5001/b:2",
			"c:3": "localhost:5001/c:3",
			"d:4": "localhost:5001/d:4",
			"e:5": "localhost:5001/e:5",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	for _, name := range []string{"a:1", "b:2", "c:3", "d:4", "e:5"} {
		assert.Contains(t, out, name)
	}
}

func TestPrintSummary_NoKeyContents(t *testing.T) {
	cfg := cli.Config{
		Host:    "host",
		KeyPath: "/home/user/.ssh/id_rsa",
		Command: "cmd",
	}
	state := &State{
		ImageMap: map[string]string{
			"a:1": "localhost:5001/a:1",
		},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.NotContains(t, out, "BEGIN")
	assert.NotContains(t, out, "PRIVATE")
	assert.NotContains(t, out, "KEY")
}
