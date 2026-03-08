package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"ship/cli"
)

// Preflight runs all preflight checks in sequence before the stage pipeline.
// Returns on first failure with a formatted error.
func Preflight(cfg cli.Config) error {
	if err := checkDocker(); err != nil {
		return err
	}
	if err := checkDockerCompose(); err != nil {
		return err
	}
	if err := checkSSH(); err != nil {
		return err
	}
	if err := checkKeyFile(cfg.KeyPath); err != nil {
		return err
	}
	if err := checkComposeFiles(cfg.ComposeFiles); err != nil {
		return err
	}
	if err := checkSSHConnectivity(cfg.KeyPath, cfg.User, cfg.Host); err != nil {
		return err
	}
	return nil
}

// checkDocker verifies docker is on PATH and responds.
func checkDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("Docker is not installed or not in PATH") //nolint:staticcheck // user-facing message per DESIGN.md spec
	}
	cmd := exec.CommandContext(context.Background(), "docker", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return errors.New("Docker is not installed or not in PATH") //nolint:staticcheck // user-facing message per DESIGN.md spec
	}
	return nil
}

// checkDockerCompose verifies docker compose V2 plugin is available.
func checkDockerCompose() error {
	cmd := exec.CommandContext(context.Background(), "docker", "compose", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker compose (V2) is required — upgrade Docker Compose or install the compose plugin")
	}
	return nil
}

// checkSSH verifies ssh is on PATH.
func checkSSH() error {
	if _, err := exec.LookPath("ssh"); err != nil {
		return errors.New("ssh is not installed or not in PATH")
	}
	return nil
}

// checkKeyFile verifies the SSH key file exists and is readable via os.Stat.
func checkKeyFile(keyPath string) error {
	if keyPath == "" {
		return fmt.Errorf("SSH key file not found: %s — verify the --key path", keyPath)
	}
	cleanPath := filepath.Clean(keyPath)
	info, err := os.Stat(cleanPath) //nolint:gosec // keyPath is user-provided CLI flag, path traversal is expected
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("SSH key file not found: %s — verify the --key path", keyPath)
		}
		return fmt.Errorf("Cannot read SSH key file: %s — check file permissions", keyPath) //nolint:staticcheck // user-facing message per DESIGN.md spec
	}
	if info.IsDir() {
		return fmt.Errorf("SSH key file not found: %s — verify the --key path", keyPath)
	}
	return nil
}

// checkComposeFiles verifies each compose file path exists.
func checkComposeFiles(paths []string) error {
	for _, p := range paths {
		cleanPath := filepath.Clean(p)
		if _, err := os.Stat(cleanPath); err != nil { //nolint:gosec // compose path is user-provided CLI flag, path traversal is expected
			if os.IsNotExist(err) {
				return fmt.Errorf("Compose file not found: %s", p) //nolint:staticcheck // user-facing message per DESIGN.md spec
			}
			return fmt.Errorf("Cannot read compose file: %s — check file permissions", p) //nolint:staticcheck // user-facing message per DESIGN.md spec
		}
	}
	return nil
}

// checkSSHConnectivity tests SSH connectivity to the remote host.
func checkSSHConnectivity(keyPath, user, host string) error {
	cmd := exec.CommandContext(context.Background(), "ssh",
		"-i", keyPath,
		"-o", "ConnectTimeout=10",
		"-o", "StrictHostKeyChecking=accept-new",
		"-o", "BatchMode=yes",
		user+"@"+host,
		"true",
	)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connection failed — verify --host and --key")
	}
	return nil
}
