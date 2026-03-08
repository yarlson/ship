# CLI Parsing & Validation

## Required Flags

- `--docker-compose` — Path(s) to Docker Compose file(s), comma-separated if multiple
- `--user` — SSH username on remote host
- `--host` — Remote host address (IP or hostname)
- `--key` — Path to SSH private key file
- `--command` — Command to execute on remote host after transfer

## Implementation

**File:** `cli/cli.go`

**Flow:**

1. `Parse(args)` creates a FlagSet with all 5 required flags
2. Parses arguments
3. Splits `--docker-compose` value on commas, trims whitespace, filters empty strings into `[]string`
4. Validates all required flags are present and non-empty
5. Detects explicit empty `--command ""` (distinct from missing `--command`)
6. Returns typed `Config` struct or error listing missing flags

**Config struct:**

```go
type Config struct {
    ComposeFiles []string // Parsed compose file paths
    User         string   // SSH user
    Host         string   // Remote host
    KeyPath      string   // SSH key path
    Command      string   // Remote command
}
```

## Help Text

Help text is a constant with usage and examples. Displayed on `--help` flag.

**Example invocations:**

Single compose file:

```bash
ship --docker-compose docker-compose.yml \
     --user deploy \
     --host 10.0.0.5 \
     --key ~/.ssh/id_ed25519 \
     --command "docker compose up -d"
```

Multiple compose files (comma-separated):

```bash
ship --docker-compose compose.yml,compose.prod.yml \
     --user root \
     --host staging.example.com \
     --key ./key.pem \
     --command "docker compose pull && docker compose up -d"
```

## Error Handling

- Missing flags → error message listing all missing flags by name + usage line
- `--help` → returns `flag.ErrHelp` (handled in main)
- No secrets logged in error messages

## Testing

Tests in `cli/cli_test.go` verify:

- Valid Config accepted
- Missing flags detected and reported
- Help flag handling
