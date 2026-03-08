package workflow

import (
	"ship/cli"
	"ship/progress"
	"ship/stage"
)

type stubStage struct {
	startMsg    string
	completeMsg string
}

var stubs = []stubStage{
	{"Starting local registry", "Registry ready"},
	{"Pushing images to local registry", "Push complete"},
	{"Establishing tunnel", "Tunnel established"},
	{"Pulling and restoring images on remote host", "Pull and restore complete"},
	{"Running remote command", "Command complete"},
}

// Run executes the 7-stage workflow.
// Stages 1-2 are real implementations; stages 3-7 remain stubs.
func Run(cfg cli.Config) error {
	// Stage 1: Build
	imageMap, err := stage.Build(cfg.ComposeFiles)
	if err != nil {
		return err
	}

	// Stage 2: Tag
	if err := stage.Tag(imageMap); err != nil {
		return err
	}

	// Stages 3-7: stubs
	for i, s := range stubs {
		n := i + 3
		progress.StageStart(n, s.startMsg)
		progress.StageComplete(n, s.completeMsg)
	}

	return nil
}
