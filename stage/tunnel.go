package stage

import (
	"context"
	"fmt"
	"time"

	"ship/cli"
	"ship/progress"
	"ship/ssh"
)

// Tunnel executes Stage 4: establish a reverse SSH tunnel to the remote host.
// Returns the tunnel process handle for lifecycle management.
func Tunnel(ctx context.Context, cfg cli.Config) (*ssh.TunnelProcess, error) {
	progress.StageStart(4, fmt.Sprintf("Establishing tunnel to %s", cfg.Host))

	tp, err := ssh.StartTunnel(ctx, cfg.KeyPath, cfg.Port, cfg.User, cfg.Host)
	if err != nil {
		return nil, fmt.Errorf("SSH tunnel failed — verify the target and SSH credentials")
	}

	// Wait for tunnel to establish, checking if the process exits early (connection failure).
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()

	select {
	case <-tp.Exited():
		// Process exited during setup — tunnel failed.
		return nil, fmt.Errorf("SSH tunnel failed — verify the target and SSH credentials")
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		// Process still alive after 2s — tunnel is established.
	}

	progress.StageComplete(4, "Tunnel established")
	return tp, nil
}
