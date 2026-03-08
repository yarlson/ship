package ssh

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"syscall"
	"time"
)

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
// Returns stdout, stderr, exit code, and error.
func RunRemoteCommand(keyPath string, port int, user, host, cmd string) (stdoutStr, stderrStr string, exitCode int, err error) {
	args := BuildRemoteCommandArgs(keyPath, port, user, host, cmd)
	c := exec.CommandContext(context.Background(), "ssh", args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	c.Stdout = &stdoutBuf
	c.Stderr = &stderrBuf

	err = c.Run()
	if err != nil {
		exitErr := &exec.ExitError{}
		if errors.As(err, &exitErr) {
			return stdoutBuf.String(), stderrBuf.String(), exitErr.ExitCode(), fmt.Errorf("remote command exited with code %d", exitErr.ExitCode())
		}
		return "", "", -1, fmt.Errorf("ssh command failed: %w", err)
	}

	return stdoutBuf.String(), stderrBuf.String(), 0, nil
}

// TunnelProcess wraps an SSH tunnel background process with lifecycle management.
type TunnelProcess struct {
	cmd  *exec.Cmd
	done chan struct{} // closed when cmd.Wait() returns
}

// Exited returns a channel that is closed when the tunnel process exits.
func (t *TunnelProcess) Exited() <-chan struct{} {
	return t.done
}

// StartTunnel starts a reverse SSH tunnel as a background process.
// Returns a TunnelProcess for lifecycle management.
func StartTunnel(keyPath string, port int, user, host string) (*TunnelProcess, error) {
	args := BuildTunnelArgs(keyPath, port, user, host)
	c := exec.CommandContext(context.Background(), "ssh", args...)

	if err := c.Start(); err != nil {
		return nil, fmt.Errorf("failed to start SSH tunnel: %w", err)
	}

	tp := &TunnelProcess{
		cmd:  c,
		done: make(chan struct{}),
	}
	go func() {
		c.Wait() //nolint:errcheck // exit status handled via Exited channel
		close(tp.done)
	}()

	return tp, nil
}

// StopTunnel gracefully stops the tunnel process.
// Sends SIGTERM first, waits with timeout, then SIGKILL if needed.
func StopTunnel(tp *TunnelProcess) error {
	if tp == nil || tp.cmd == nil || tp.cmd.Process == nil {
		return nil
	}

	// Check if already exited.
	select {
	case <-tp.done:
		return nil
	default:
	}

	// Send SIGTERM.
	if err := tp.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// Process already exited.
		return nil //nolint:nilerr // process already exited
	}

	// Wait with timeout.
	select {
	case <-tp.done:
		return nil
	case <-time.After(5 * time.Second):
		// Force kill.
		_ = tp.cmd.Process.Kill() //nolint:errcheck // best-effort
		<-tp.done
		return nil
	}
}
