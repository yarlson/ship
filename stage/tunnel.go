package stage

import (
	"fmt"
	"time"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Tunnel executes Stage 5: establish a reverse SSH tunnel to the remote host.
// Returns the tunnel process handle for lifecycle management.
func Tunnel(cfg cli.Config) (*ssh.TunnelProcess, error) {
	progress.StageStart(5, fmt.Sprintf("Establishing tunnel to %s", cfg.Host))

	tp, err := ssh.StartTunnel(cfg.KeyPath, cfg.User, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("SSH tunnel failed — connection refused (verify --host and --key)")
	}

	// Wait for tunnel to establish, checking if the process exits early (connection failure).
	select {
	case <-tp.Exited():
		// Process exited during setup — tunnel failed.
		return nil, fmt.Errorf("SSH tunnel failed — connection refused (verify --host and --key)")
	case <-time.After(2 * time.Second):
		// Process still alive after 2s — tunnel is established.
	}

	progress.StageComplete(5, "Tunnel established")
	return tp, nil
}
