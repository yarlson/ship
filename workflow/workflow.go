package workflow

import (
	"fmt"
	"sort"
	"strings"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
	"ship/stage"
)

// State holds shared mutable state passed through stages.
type State struct {
	ImageMap  map[string]string
	TunnelCmd *ssh.TunnelProcess
}

// Run executes preflight checks followed by the 7-stage workflow.
func Run(cfg cli.Config) error {
	if err := Preflight(cfg); err != nil {
		return err
	}

	state := &State{}

	// Stage 1: Build
	imageMap, err := stage.Build(cfg.ComposeFiles)
	if err != nil {
		return wrapStageErr(1, "Build", err)
	}
	state.ImageMap = imageMap

	// Stage 2: Tag
	if err := stage.Tag(imageMap); err != nil {
		return wrapStageErr(2, "Tag", err)
	}

	// Stage 3: Registry
	if err := stage.Registry(); err != nil {
		return wrapStageErr(3, "Registry", err)
	}

	// Stage 4: Push
	if err := stage.Push(imageMap); err != nil {
		return wrapStageErr(4, "Push", err)
	}

	// Stage 5: Tunnel
	tp, err := stage.Tunnel(cfg)
	if err != nil {
		return wrapStageErr(5, "Tunnel", err)
	}
	state.TunnelCmd = tp
	defer cleanupTunnel(state)

	// Stage 6: Pull & Restore
	if err := stage.Pull(cfg, imageMap); err != nil {
		return wrapStageErr(6, "Pull", err)
	}

	// Stage 7: Command
	if err := stage.Command(cfg); err != nil {
		return wrapStageErr(7, "Command", err)
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
	names := make([]string, 0, len(state.ImageMap))
	for original := range state.ImageMap {
		names = append(names, original)
	}
	sort.Strings(names)

	fmt.Fprintln(progress.Writer, "Ship complete")
	fmt.Fprintf(progress.Writer, "  Host:     %s\n", cfg.Host)
	fmt.Fprintf(progress.Writer, "  Images:   %s\n", strings.Join(names, ", "))
	fmt.Fprintf(progress.Writer, "  Command:  %s\n", cfg.Command)
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
