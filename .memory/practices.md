# Practices & Conventions

## Code

- Shell out to `docker` and `ssh` CLIs — no Go SDK libraries
- Use `exec.CommandContext` (not `exec.Command`) to satisfy `noctx` linter
- Named return values on multi-return functions (gocritic/unnamedResult)
- Preallocate slices when capacity is known
- No stutter in exported type names (`workflow.State` not `workflow.WorkflowState`)
- `progress.Writer` is the single output destination for all user-facing messages

## Testing

- Pure TDD, outside-in: failing test first, then minimal code
- Unit tests: no build tags, test argument building and output formatting
- Integration tests: `//go:build integration` tag, require Docker and SSH access
- Test server: root@46.101.213.82, key ~/.ssh/id_rsa
- Cross-arch testing: compose files must specify `platform: linux/amd64` and `build.platforms: [linux/amd64]` (arm64 Mac → amd64 server)
- Capture progress output by swapping `progress.Writer` to `bytes.Buffer`
- `testlock.Port5001(t)` for mutex on port 5001 across parallel tests

## SSH Tunnel Lifecycle

- `TunnelProcess` struct owns a single `cmd.Wait()` goroutine with a `done` channel
- Prevents data race from multiple `Wait()` calls
- `StopTunnel`: SIGTERM → 5s timeout → SIGKILL → wait on done channel
- Tunnel health check: 2s `select` on `Exited()` channel after start

## Testing Tunnel Failures

- Use a bad key file (causes instant SSH auth failure) instead of unreachable IP (hangs past ConnectTimeout=5s)

## Lint

- `golangci-lint run --fix ./...` must pass clean before finishing
- Key rules: noctx, errcheck, gocritic (appendAssign, unnamedResult), prealloc, revive
