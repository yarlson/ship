package workflow

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"ship/cli"
	"ship/docker"
	shipssh "ship/ssh"
)

// Preflight runs all preflight checks in sequence before the stage pipeline.
// Returns on first failure with a formatted error.
func Preflight(cfg cli.Config) error {
	if err := checkDocker(); err != nil {
		return err
	}
	if err := checkSSH(); err != nil {
		return err
	}
	if err := checkKeyFile(cfg.KeyPath); err != nil {
		return err
	}
	if err := checkLocalImages(cfg.Images); err != nil {
		return err
	}
	if err := checkSSHConnectivity(cfg); err != nil {
		return err
	}
	return nil
}

// checkDocker verifies docker is on PATH and responds.
func checkDocker() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker is not installed or not in PATH")
	}
	cmd := exec.CommandContext(context.Background(), "docker", "version")
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return errors.New("docker is not installed or not in PATH")
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
// Empty means the user chose the default SSH identity behavior.
func checkKeyFile(keyPath string) error {
	if keyPath == "" {
		return nil
	}

	cleanPath := filepath.Clean(keyPath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("SSH key file not found: %s — verify the -i path", keyPath)
		}
		return fmt.Errorf("cannot read SSH key file: %s — check file permissions", keyPath)
	}
	if info.IsDir() {
		return fmt.Errorf("SSH key file not found: %s — verify the -i path", keyPath)
	}
	return nil
}

func checkLocalImages(imageRefs []string) error {
	for _, imageRef := range imageRefs {
		if err := docker.ImageExists(imageRef); err != nil {
			return err
		}
	}

	return nil
}

// checkSSHConnectivity tests SSH connectivity to the remote host.
func checkSSHConnectivity(cfg cli.Config) error {
	args := shipssh.BuildRemoteCommandArgs(cfg.KeyPath, cfg.Port, cfg.User, cfg.Host, "true")
	cmd := exec.CommandContext(context.Background(), "ssh", args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connection failed — verify the target and SSH credentials")
	}
	return nil
}
