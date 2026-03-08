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
	{"Establishing tunnel", "Tunnel established"},
	{"Pulling and restoring images on remote host", "Pull and restore complete"},
	{"Running remote command", "Command complete"},
}

// Run executes the 7-stage workflow.
// Stages 1-4 are real implementations; stages 5-7 remain stubs.
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

	// Stage 3: Registry
	if err := stage.Registry(); err != nil {
		return err
	}

	// Stage 4: Push
	if err := stage.Push(imageMap); err != nil {
		return err
	}

	// Stages 5-7: stubs
	for i, s := range stubs {
		n := i + 5
		progress.StageStart(n, s.startMsg)
		progress.StageComplete(n, s.completeMsg)
	}

	return nil
}
