package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

// RemoteCommandResult is the captured output from one SSH remote command.
type RemoteCommandResult struct {
	Stdout string
	Stderr string
}

// RemoteCommandError reports a non-zero exit from a remote command.
type RemoteCommandError struct {
	ExitCode int
}

func (e *RemoteCommandError) Error() string {
	return fmt.Sprintf("remote command exited with code %d", e.ExitCode)
}

// BuildRemoteCommandArgs returns the argument slice for an SSH remote command execution.
func BuildRemoteCommandArgs(keyPath string, port int, user, host, cmd string) []string {
	args := commonArgs(keyPath, port)
	args = append(args,
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		user+"@"+host,
		cmd,
	)
	return args
}

// BuildTunnelArgs returns the argument slice for a reverse SSH tunnel.
func BuildTunnelArgs(keyPath string, port int, user, host string) []string {
	args := commonArgs(keyPath, port)
	args = append(args,
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		"-o", "ExitOnForwardFailure=yes",
		"-R", "5001:localhost:5001",
		"-N",
		user+"@"+host,
	)
	return args
}

func commonArgs(keyPath string, port int) []string {
	args := make([]string, 0, 4)
	if keyPath != "" {
		args = append(args, "-i", keyPath)
	}
	if port > 0 {
		args = append(args, "-p", strconv.Itoa(port))
	}
	return args
}

// RunRemoteCommand executes a command on the remote host via SSH.
func RunRemoteCommand(ctx context.Context, keyPath string, port int, user, host, cmd string) (RemoteCommandResult, error) {
	args := BuildRemoteCommandArgs(keyPath, port, user, host, cmd)
	c := exec.CommandContext(ctx, "ssh", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	c.Stdout = &stdoutBuf
	c.Stderr = &stderrBuf

	err := c.Run()
	result := RemoteCommandResult{
		Stdout: stdoutBuf.String(),
		Stderr: stderrBuf.String(),
	}
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return result, err
		}

		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return result, &RemoteCommandError{ExitCode: exitErr.ExitCode()}
		}
		return RemoteCommandResult{}, fmt.Errorf("ssh command failed: %w", err)
	}

	return result, nil
}

// TunnelProcess wraps an SSH tunnel background process with lifecycle management.
type TunnelProcess struct {
	cmd    *exec.Cmd
	done   chan struct{} // closed when cmd.Wait() returns
	signal func(os.Signal) error
	kill   func() error
}

// Exited returns a channel that is closed when the tunnel process exits.
func (t *TunnelProcess) Exited() <-chan struct{} {
	return t.done
}

// StartTunnel starts a reverse SSH tunnel as a background process.
// Returns a TunnelProcess for lifecycle management.
func StartTunnel(ctx context.Context, keyPath string, port int, user, host string) (*TunnelProcess, error) {
	args := BuildTunnelArgs(keyPath, port, user, host)
	c := exec.CommandContext(ctx, "ssh", args...)

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("failed to start SSH tunnel: %w", err)
	}

	tp := &TunnelProcess{
		cmd:    c,
		done:   make(chan struct{}),
		signal: c.Process.Signal,
		kill:   c.Process.Kill,
	}
	go func() {
		c.Wait() //nolint:errcheck // exit status handled via Exited channel
		close(tp.done)
	}()

	return tp, nil
}

// StopTunnel gracefully stops the tunnel process.
// Sends SIGTERM first, waits with timeout, then SIGKILL if needed.
func StopTunnel(ctx context.Context, tp *TunnelProcess) error {
	if tp == nil || tp.cmd == nil || tp.cmd.Process == nil {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	// Check if already exited.
	select {
	case <-tp.done:
		return nil
	default:
	}

	// Send SIGTERM.
	if err := tp.signalProcess(syscall.SIGTERM); err != nil {
		// Process already exited.
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return fmt.Errorf("failed to stop SSH tunnel: %w", err)
	}

	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()

	select {
	case <-tp.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		// Force kill.
		if err := tp.killProcess(); err != nil && !errors.Is(err, os.ErrProcessDone) {
			return fmt.Errorf("failed to force-stop SSH tunnel: %w", err)
		}
		select {
		case <-tp.done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (t *TunnelProcess) signalProcess(sig os.Signal) error {
	if t.signal != nil {
		return t.signal(sig)
	}

	return t.cmd.Process.Signal(sig)
}

func (t *TunnelProcess) killProcess() error {
	if t.kill != nil {
		return t.kill()
	}

	return t.cmd.Process.Kill()
}
