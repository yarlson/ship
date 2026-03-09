package stage

import (
	"context"
	"fmt"

	"ship/docker"
	"ship/progress"
)

// Registry executes Stage 2: ensure a local registry is running on :5001.
func Registry(ctx context.Context) error {
	progress.StageStart(2, "Starting local registry")

	running, err := docker.CheckRegistryRunning(ctx)
	if err != nil {
		return fmt.Errorf("Failed to check registry status — %w", err) //nolint:staticcheck // user-facing message per DESIGN.md spec
	}

	if !running {
		conflict, err := docker.CheckPortConflict(ctx)
		if err != nil {
			return fmt.Errorf("Failed to check port 5001 — %w", err) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}
		if conflict {
			return fmt.Errorf("Port 5001 already in use — stop the existing process or free the port") //nolint:staticcheck // user-facing message per DESIGN.md spec
		}

		if err := docker.StartRegistry(ctx); err != nil {
			return fmt.Errorf("Failed to start registry — %w", err) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}
	}

	progress.StageComplete(2, "Registry ready")
	return nil
}
