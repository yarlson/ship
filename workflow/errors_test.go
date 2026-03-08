package workflow

import (
	"errors"
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStageError_FormatWithHint(t *testing.T) {
	se := &StageError{
		Stage: 5,
		Name:  "Tunnel",
		Err:   errors.New("connection refused"),
		Hint:  "verify --host and --key",
	}

	assert.Equal(t, "connection refused — verify --host and --key", se.Error())
}

func TestStageError_FormatWithoutHint(t *testing.T) {
	se := &StageError{
		Stage: 1,
		Name:  "Build",
		Err:   errors.New("build failed"),
		Hint:  "",
	}

	assert.Equal(t, "build failed", se.Error())
}

func TestStageError_Unwrap(t *testing.T) {
	sentinel := errors.New("sentinel")
	se := &StageError{
		Stage: 3,
		Name:  "Registry",
		Err:   sentinel,
		Hint:  "check registry",
	}

	require.True(t, errors.Is(se, sentinel))
}

func TestStageError_NoKeyContents(t *testing.T) {
	se := &StageError{
		Stage: 5,
		Name:  "Tunnel",
		Err:   errors.New("failed to connect using /home/user/.ssh/id_rsa"),
		Hint:  "verify --key path",
	}

	msg := se.Error()
	assert.Contains(t, msg, "/home/user/.ssh/id_rsa")
	assert.NotContains(t, msg, "BEGIN")
	assert.NotContains(t, msg, "PRIVATE")
	assert.NotContains(t, msg, "KEY")
}

func TestStageError_NoEmoji(t *testing.T) {
	se := &StageError{
		Stage: 1,
		Name:  "Build",
		Err:   errors.New("build failed"),
		Hint:  "see docker compose output above",
	}

	emoji := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{1F1E0}-\x{1F1FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	assert.False(t, emoji.MatchString(se.Error()))
}

func TestStageError_NoANSI(t *testing.T) {
	se := &StageError{
		Stage: 1,
		Name:  "Build",
		Err:   errors.New("build failed"),
		Hint:  "see docker compose output above",
	}

	assert.NotContains(t, se.Error(), "\x1b[")
}

func TestStageError_NoHedging(t *testing.T) {
	se := &StageError{
		Stage: 1,
		Name:  "Build",
		Err:   errors.New("build failed"),
		Hint:  "see docker compose output above",
	}

	msg := se.Error()
	assert.NotContains(t, msg, "might")
	assert.NotContains(t, msg, "possibly")
	assert.NotContains(t, msg, "try to")
	assert.NotContains(t, msg, "maybe")
}
