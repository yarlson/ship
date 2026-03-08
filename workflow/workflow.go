package workflow

import (
	"fmt"
	"strings"

	"ship/cli"
	"ship/docker"
	"ship/progress"
	"ship/ssh"
	"ship/stage"
)

// State holds shared mutable state passed through stages.
type State struct {
	OriginalImages []string
	TransferImages []string
	TunnelCmd      *ssh.TunnelProcess
}

// Run executes preflight checks followed by the 5-stage workflow.
func Run(cfg cli.Config) error {
	if err := Preflight(cfg); err != nil {
		return err
	}

	state := &State{
		OriginalImages: append([]string(nil), cfg.Images...),
		TransferImages: docker.TransferTags(cfg.Images),
	}

	if err := stage.Tag(state.OriginalImages, state.TransferImages); err != nil {
		return wrapStageErr(1, "Tag", err)
	}

	if err := stage.Registry(); err != nil {
		return wrapStageErr(2, "Registry", err)
	}

	if err := stage.Push(state.TransferImages); err != nil {
		return wrapStageErr(3, "Push", err)
	}

	tp, err := stage.Tunnel(cfg)
	if err != nil {
		return wrapStageErr(4, "Tunnel", err)
	}
	state.TunnelCmd = tp
	defer cleanupTunnel(state)

	if err := stage.Pull(cfg, state.OriginalImages, state.TransferImages); err != nil {
		return wrapStageErr(5, "Pull", err)
	}

	printSummary(cfg, state)
	return nil
}

// wrapStageErr wraps a stage error in StageError for consistent formatting.
func wrapStageErr(stageNum int, name string, err error) *StageError {
	return &StageError{
		Stage: stageNum,
		Name:  name,
		Err:   err,
	}
}

// printSummary prints the success summary block to stdout.
func printSummary(cfg cli.Config, state *State) {
	fmt.Fprintln(progress.Writer, "Ship complete")
	fmt.Fprintf(progress.Writer, "  Host:     %s\n", cfg.Host)
	fmt.Fprintf(progress.Writer, "  Images:   %s\n", strings.Join(state.OriginalImages, ", "))
	fmt.Fprintln(progress.Writer, "  Status:   Success")
}

// cleanupTunnel stops the SSH tunnel process if it is running.
func cleanupTunnel(state *State) {
	if state.TunnelCmd == nil {
		return
	}
	if err := ssh.StopTunnel(state.TunnelCmd); err != nil {
		fmt.Fprintf(progress.Writer, "Warning: tunnel cleanup failed: %s\n", err)
	}
}
