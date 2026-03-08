# SSH Tunnel and Remote Command Execution

**File:** `ssh/ssh.go`

## Overview

The `ssh` module encapsulates all SSH operations: reverse tunnel establishment, remote command execution, and tunnel lifecycle management. No Go SSH libraries used — shells out to `ssh` CLI directly.

## Key Functions

### RemoteCommand: RunRemoteCommand

**Function:** `RunRemoteCommand(keyPath, user, host, cmd string) (stdoutStr, stderrStr string, exitCode int, err error)`

Executes a command on a remote host and returns output.

#### Flow

1. Build SSH command args via `BuildRemoteCommandArgs()`
2. Execute `ssh -i <key> -o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=5 user@host cmd`
3. Capture stdout and stderr to buffers
4. Run command and wait for completion
5. Extract exit code from ExitError if present
6. Return (stdout, stderr, exitCode, error)

#### SSH Options

- `-i <key>` — Use private key for authentication
- `-o StrictHostKeyChecking=no` — Skip host key check (safe for one-time CLI use)
- `-o BatchMode=yes` — No password prompt (fail if key fails)
- `-o ConnectTimeout=5` — 5-second connection timeout

#### Error Handling

- **SSH command fails** — `ssh command failed: <error>`
  - Wraps exec error (network error, key not found, etc.)
  - Distinguishes from command exit code

- **Command exits non-zero** — Returns exitCode > 0
  - Caller interprets as failure
  - stdout/stderr still captured and returned

#### Contract

- **Input:** SSH credentials, remote command
- **Output:** Command stdout, stderr, exit code, or SSH error
- **No I/O redirection** — Doesn't read stdin, allows TTY output

### Tunnel: StartTunnel

**Function:** `StartTunnel(keyPath, user, host string) (*TunnelProcess, error)`

Starts a reverse SSH tunnel as a background process.

#### Flow

1. Build tunnel args via `BuildTunnelArgs()`
2. Execute `ssh -i <key> -o StrictHostKeyChecking=no -o BatchMode=yes -o ConnectTimeout=5 -o ExitOnForwardFailure=yes -R 5001:localhost:5001 -N user@host`
3. Start process in background with `cmd.Start()`
4. Wrap in TunnelProcess struct with `done` channel
5. Spawn goroutine to wait on process: `c.Wait()` then close `done` channel
6. Return (TunnelProcess, nil) immediately without blocking
7. Process continues running in background

#### SSH Tunnel Options

- `-R 5001:localhost:5001` — Reverse forward: remote port 5001 → local port 5001
- `-N` — Don't execute command (tunnel only)
- `-o ExitOnForwardFailure=yes` — Exit if port forwarding fails (e.g., port already in use on remote)

#### Error Handling

- **Start fails** — `failed to start SSH tunnel: <error>`
  - Wraps exec error (key not found, host not found, etc.)
  - Process never started

#### Contract

- **Input:** SSH credentials, target host
- **Output:** TunnelProcess handle for monitoring and cleanup, or error
- **Side effect:** Reverse SSH tunnel established and running in background
- **No blocking** — Returns immediately; process runs in background

### TunnelProcess: Lifecycle Management

**Type:** `struct TunnelProcess`

```go
type TunnelProcess struct {
    cmd  *exec.Cmd
    done chan struct{}  // closed when cmd.Wait() returns
}
```

#### Exited() Method

**Function:** `(t *TunnelProcess) Exited() <-chan struct{}`

Returns a channel that closes when the tunnel process exits. Used to detect early failure.

#### Contract

- **Input:** None
- **Output:** Read-only channel
- **Semantics:** Select on this channel to wait for process exit

### Tunnel: StopTunnel

**Function:** `StopTunnel(tp *TunnelProcess) error`

Gracefully stops a running tunnel process with SIGTERM → SIGKILL fallback.

#### Flow

1. Check if TunnelProcess is nil or already exited (select on `done`)
2. Send SIGTERM to process
3. Wait up to 5 seconds for graceful shutdown
4. If still running after 5s, send SIGKILL (best-effort)
5. Wait for process to exit
6. Return nil (always succeeds or process already gone)

#### Invariants

- **Safe to call multiple times** — Checks if already exited before sending signal
- **Safe with nil** — Returns nil if TunnelProcess is nil
- **Always succeeds** — Returns nil even if signals fail (process may have exited)
- **Non-blocking** — Doesn't block indefinitely, has 5-second timeout

#### Error Handling

- **Process already exited** — Returns nil (no error)
- **Signal fails** — Returns nil (process was already exiting)
- **SIGKILL fails** — Still returns nil (best-effort, process likely exiting anyway)

## Data Race Prevention

**Pattern:** Single goroutine owns `cmd.Wait()`

The `TunnelProcess` design prevents data races:

1. `StartTunnel()` spawns **one** goroutine to call `cmd.Wait()`
2. That goroutine closes the `done` channel when `cmd.Wait()` returns
3. Other code only **reads** from `done` channel (select, waiting)
4. No shared mutable state — only one reader of exit status

This avoids multiple goroutines calling `cmd.Wait()` simultaneously (undefined behavior in Go).

## SSH Options Strategy

All SSH connections use the same safety-focused options:

- `StrictHostKeyChecking=no` — Skip host key check (acceptable for CLI tools run by user)
- `BatchMode=yes` — Non-interactive (fail immediately if auth fails, no password prompt)
- `ConnectTimeout=5` — Timeout bad connections quickly (don't hang)

Rationale: Ship runs as CLI tool in user's environment. User provides host, key, command. No need for interactive prompts or strict host validation (equivalent to `ssh` user would run manually).

## Integration with Stages

- **Stage 5 (Tunnel):** Calls `StartTunnel()`, monitors `Exited()`, returns TunnelProcess
- **Stage 6 (Pull):** Calls `RunRemoteCommand()` for `docker pull` and `docker tag`
- **Stage 7 (Command):** Calls `RunRemoteCommand()` for user-provided command
- **Cleanup:** Deferred `StopTunnel()` stops tunnel after all stages

## Testing

- **ssh_test.go** — Unit tests for SSH arg building
- **ssh_integration_test.go** — Integration tests against real SSH connections and test server
