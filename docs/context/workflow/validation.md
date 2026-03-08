# Preflight Validation & Error Handling

## Preflight Validation

Preflight checks run at the start of `workflow.Run()` before any stages execute. Function: `workflow.Preflight(cfg)`.

**Check sequence:**

1. **Docker installed** — `docker version` returns 0
   - Error: `Docker is not installed or not in PATH`
   - No hint (essential dependency, no workaround)

2. **Docker Compose V2** — `docker compose version` returns 0
   - Error: `docker compose (V2) is required — upgrade Docker Compose or install the compose plugin`
   - Hint included in error message

3. **SSH installed** — `ssh` command found in PATH via `exec.LookPath("ssh")`
   - Error: `ssh is not installed or not in PATH`
   - No hint (essential dependency)

4. **SSH key file accessible** — `os.Stat(filepath.Clean(keyPath))` succeeds
   - Error on not found: `SSH key file not found: <path> — verify the --key path`
   - Error on unreadable: `Cannot read SSH key file: <path> — check file permissions`
   - Hint included in error message

5. **SSH connectivity** — `ssh -i <key> -o ConnectTimeout=10 -o StrictHostKeyChecking=accept-new -o BatchMode=yes <user>@<host> true` returns 0
   - Error: `SSH connection failed — verify --host and --key`
   - Hint included in error message

**Exit behavior:** Returns on first failure. No stages execute if any check fails.

## StageError Type

**File:** `workflow/errors.go`

**Structure:**

```go
type StageError struct {
    Stage int    // 1-7
    Name  string // "Build", "Tag", etc.
    Err   error  // underlying error
    Hint  string // optional hint text
}

// Error() returns "<what> — <hint>" if hint present, else "<what>"
// Unwrap() returns Err for errors.Is/errors.As compatibility
```

**Usage:**

- All stage failures wrapped: `return wrapStageErr(stageNum, stageName, err)`
- Preflight errors returned directly without wrapping
- `main.go` adds `Error: ` prefix when printing to stderr

**Format examples:**

- With hint: `Error: SSH connection failed — verify --host and --key`
- Without hint: `Error: Cannot read SSH key file: /path/to/key — check file permissions`

## Fail-Fast Pattern

- Preflight checks run in sequence, exit on first failure
- Stages execute in sequence, exit on first failure
- No recovery or retries within workflow
- Errors bubble up to `main.go` which prints and exits with code 1
