// Package testlock provides file-based locking for integration tests
// that need exclusive access to shared resources like port 5001.
package testlock

import (
	"context"
	"net"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Port5001 acquires an exclusive file lock to serialize access to port 5001
// across test packages. The lock is released when the test completes.
func Port5001(t *testing.T) {
	t.Helper()

	lockFile := os.TempDir() + "/ship-test-port5001.lock"
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("open lock file: %v", err)
	}

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		t.Fatalf("acquire lock: %v", err)
	}

	t.Cleanup(func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:errcheck // best-effort unlock
		f.Close()
	})
}

// StopRegistry stops any running registry:2 containers on port 5001 and waits for port release.
func StopRegistry(t *testing.T) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), "docker", "ps", "-q", "--filter", "ancestor=registry:2", "--filter", "publish=5001")
	out, err := cmd.Output()
	if err != nil {
		return
	}
	ids := strings.TrimSpace(string(out))
	if ids == "" {
		return
	}
	for _, id := range strings.Split(ids, "\n") {
		id = strings.TrimSpace(id)
		if id != "" {
			//nolint:errcheck // best-effort cleanup in tests
			exec.CommandContext(context.Background(), "docker", "rm", "-f", id).Run()
		}
	}
	WaitPort5001Free(t)
}

// WaitPort5001Free waits until port 5001 is no longer accepting connections.
func WaitPort5001Free(t *testing.T) {
	t.Helper()
	dialer := &net.Dialer{Timeout: 200 * time.Millisecond}
	for range 30 {
		conn, err := dialer.DialContext(context.Background(), "tcp", "localhost:5001")
		if err != nil {
			return
		}
		conn.Close()
		time.Sleep(100 * time.Millisecond)
	}
}
