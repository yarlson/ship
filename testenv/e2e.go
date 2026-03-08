package testenv

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

// E2EConfig holds the SSH target used by end-to-end tests.
type E2EConfig struct {
	User    string
	Host    string
	KeyPath string
}

// Address returns the SSH destination in user@host form.
func (c E2EConfig) Address() string {
	return c.User + "@" + c.Host
}

// LoadE2EConfigFromEnv reads E2E test configuration from environment variables.
func LoadE2EConfigFromEnv() (E2EConfig, error) {
	cfg := E2EConfig{
		User:    strings.TrimSpace(os.Getenv("SHIP_E2E_USER")),
		Host:    strings.TrimSpace(os.Getenv("SHIP_E2E_HOST")),
		KeyPath: strings.TrimSpace(os.Getenv("SHIP_E2E_KEY")),
	}

	var missing []string
	if cfg.User == "" {
		missing = append(missing, "SHIP_E2E_USER")
	}
	if cfg.Host == "" {
		missing = append(missing, "SHIP_E2E_HOST")
	}
	if len(missing) > 0 {
		return E2EConfig{}, fmt.Errorf("missing E2E test configuration: set %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}

// RequireE2EConfig loads and validates E2E test configuration, skipping the test
// when required runtime prerequisites are unavailable.
func RequireE2EConfig(t *testing.T) E2EConfig {
	t.Helper()

	cfg, err := LoadE2EConfigFromEnv()
	if err != nil {
		t.Skipf("skipping e2e test: %v", err)
	}

	dockerCmd := exec.CommandContext(context.Background(), "docker", "version")
	if err := dockerCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: Docker daemon unavailable: %v", err)
	}

	if cfg.KeyPath != "" {
		if _, err := os.Stat(cfg.KeyPath); err != nil {
			t.Skipf("skipping e2e test: SSH key unavailable: %v", err)
		}
	}

	args := []string{}
	if cfg.KeyPath != "" {
		args = append(args, "-i", cfg.KeyPath)
	}
	args = append(args,
		"-p", strconv.Itoa(22),
		"-o", "StrictHostKeyChecking=no",
		"-o", "BatchMode=yes",
		"-o", "ConnectTimeout=5",
		cfg.Address(),
		"true",
	)
	sshCmd := exec.CommandContext(context.Background(), "ssh", args...)
	if err := sshCmd.Run(); err != nil {
		t.Skipf("skipping e2e test: SSH test host unavailable: %v", err)
	}

	return cfg
}
