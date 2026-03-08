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
		Host:   "h.example.com",
		Images: []string{"app:latest", "traefik:v3"},
	}
	state := &State{
		OriginalImages: []string{"app:latest", "traefik:v3"},
		TransferImages: []string{"localhost:5001/app:latest", "localhost:5001/traefik:v3"},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "Ship complete")
	assert.Contains(t, out, "Host:     h.example.com")
	assert.Contains(t, out, "Images:   app:latest, traefik:v3")
	assert.Contains(t, out, "Status:   Success")
}

func TestPrintSummary_SingularLabel(t *testing.T) {
	cfg := cli.Config{
		Host:   "h.example.com",
		Images: []string{"app:latest"},
	}
	state := &State{
		OriginalImages: []string{"app:latest"},
		TransferImages: []string{"localhost:5001/app:latest"},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "Image:    app:latest")
	assert.NotContains(t, out, "Images:")
}

func TestPrintSummary_LabelAlignment(t *testing.T) {
	cfg := cli.Config{
		Host:   "host",
		Images: []string{"app:latest"},
	}
	state := &State{
		OriginalImages: []string{"app:latest"},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	lines := strings.Split(strings.TrimSpace(out), "\n")
	for _, line := range lines[1:] {
		assert.True(t, len(line) >= 12, "line too short: %s", line)
		prefix := line[:12]
		assert.True(t, strings.HasSuffix(strings.TrimRight(prefix, " "), ":"),
			"label should end with colon: %q", prefix)
	}
}

func TestPrintSummary_NoTransferTagsInOutput(t *testing.T) {
	cfg := cli.Config{
		Host:   "host",
		Images: []string{"app:latest", "traefik:v3"},
	}
	state := &State{
		OriginalImages: []string{"app:latest", "traefik:v3"},
		TransferImages: []string{"localhost:5001/app:latest", "localhost:5001/traefik:v3"},
	}

	out := captureSummaryOutput(func() {
		printSummary(cfg, state)
	})

	assert.Contains(t, out, "app:latest, traefik:v3")
	assert.NotContains(t, out, "localhost:5001/")
}
