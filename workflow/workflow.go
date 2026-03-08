package workflow

import (
	"ship/cli"
	"ship/progress"
)

type stage struct {
	startMsg    string
	completeMsg string
}

var stages = []stage{
	{"Building images", "Build complete"},
	{"Tagging images for transfer", "Tag complete"},
	{"Starting local registry", "Registry ready"},
	{"Pushing images to local registry", "Push complete"},
	{"Establishing tunnel", "Tunnel established"},
	{"Pulling and restoring images on remote host", "Pull and restore complete"},
	{"Running remote command", "Command complete"},
}

// Run executes the 7-stage workflow using stub implementations.
func Run(_ cli.Config) error {
	for i, s := range stages {
		n := i + 1
		progress.StageStart(n, s.startMsg)
		progress.StageComplete(n, s.completeMsg)
	}
	return nil
}
