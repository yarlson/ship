package docker

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"
	"time"
)

// ParseRegistryContainerFilter parses docker ps output to determine if a registry container is running.
// Returns true if the output contains at least one container ID.
func ParseRegistryContainerFilter(output string) bool {
	return strings.TrimSpace(output) != ""
}

// CheckRegistryRunning checks if a registry:2 container is running with port 5001 mapped.
func CheckRegistryRunning() (bool, error) {
	cmd := exec.CommandContext(context.Background(), "docker", "ps", "--filter", "ancestor=registry:2", "--filter", "publish=5001", "--format", "{{.ID}}")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("docker ps: %w", err)
	}
	return ParseRegistryContainerFilter(string(out)), nil
}

// CheckPortConflict detects if port 5001 is occupied by any running container (non-registry).
// Should only be called after CheckRegistryRunning returns false.
func CheckPortConflict() (bool, error) {
	// Check if any container is using port 5001.
	cmd := exec.CommandContext(context.Background(), "docker", "ps", "--filter", "publish=5001", "--format", "{{.ID}}")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("docker ps: %w", err)
	}
	if strings.TrimSpace(string(out)) != "" {
		return true, nil
	}

	// Also check non-Docker processes via TCP dial.
	dialer := &net.Dialer{Timeout: 1 * time.Second}
	conn, dialErr := dialer.DialContext(context.Background(), "tcp", "localhost:5001")
	if dialErr != nil {
		return false, nil //nolint:nilerr // connection refused means port is free, not an error
	}
	conn.Close()
	return true, nil
}

// StartRegistry starts a registry:2 container on port 5001 and waits for it to accept connections.
func StartRegistry() error {
	cmd := exec.CommandContext(context.Background(), "docker", "run", "-d", "-p", "5001:5000", "registry:2")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker run registry: %s", strings.TrimSpace(string(out)))
	}

	// Wait for the registry to be ready to accept connections.
	dialer := &net.Dialer{Timeout: 500 * time.Millisecond}
	for range 30 {
		conn, err := dialer.DialContext(context.Background(), "tcp", "localhost:5001")
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("registry started but not accepting connections on port 5001")
}

// PushImage runs docker push for the given image reference.
func PushImage(imageRef string) error {
	cmd := exec.CommandContext(context.Background(), "docker", "push", imageRef)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("docker push: %s", strings.TrimSpace(string(out)))
	}
	return nil
}
