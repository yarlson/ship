package progress

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func captureOutput(fn func()) string {
	var buf bytes.Buffer
	orig := Writer
	Writer = &buf
	defer func() { Writer = orig }()
	fn()
	return buf.String()
}

func TestStageStart_Format(t *testing.T) {
	out := captureOutput(func() {
		StageStart(1, "Tagging image for transfer")
	})

	assert.Equal(t, "[1/5] Tagging image for transfer...\n", out)
}

func TestStageComplete_Format(t *testing.T) {
	out := captureOutput(func() {
		StageComplete(2, "Registry ready")
	})

	assert.Equal(t, "[2/5] Registry ready\n", out)
}

func TestStageStart_NoANSI(t *testing.T) {
	out := captureOutput(func() {
		StageStart(1, "Tagging image for transfer")
	})

	ansi := regexp.MustCompile(`\x1b\[`)
	assert.False(t, ansi.MatchString(out), "output contains ANSI escape codes")
}

func TestStageComplete_NoEmoji(t *testing.T) {
	out := captureOutput(func() {
		StageComplete(5, "Pull and restore complete")
	})

	// Check for common emoji Unicode ranges.
	emoji := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	assert.False(t, emoji.MatchString(out), "output contains emoji")
}

func TestStageStart_AllFiveNumbers(t *testing.T) {
	for i := 1; i <= 5; i++ {
		out := captureOutput(func() {
			StageStart(i, "Test")
		})
		assert.Contains(t, out, "["+string(rune('0'+i))+"/5]")
	}
}

func TestProgress_NoBlankLines(t *testing.T) {
	out := captureOutput(func() {
		for i := 1; i <= 5; i++ {
			StageStart(i, "Start")
			StageComplete(i, "Done")
		}
	})

	assert.NotContains(t, out, "\n\n", "output contains blank lines")
}
