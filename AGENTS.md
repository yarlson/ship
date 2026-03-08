Ship is a Go CLI tool that builds Docker Compose images locally, transfers them to a remote host via SSH tunnel and local registry, and runs a deployment command — all in one invocation.

## Commands

```bash
# Build
go build -o ship .

# Lint (includes formatting via gci/gofmt/gofumpt/goimports)
golangci-lint run --fix ./...

# Test
go test -race -count=1 -v -timeout=120s ./...

# Integration tests (require Docker)
go test -race -count=1 -v -timeout=120s -tags=integration ./...

# Format docs
bunx prettier --write "**/*.md"
```

## Principles

- DRY, KISS, SOLID, YAGNI — in that priority order
- Pure TDD, outside-in: write the failing test first, then the minimal code to pass it
- Shell out to `docker` and `ssh` CLIs — no Go SDK libraries for these
- Fail fast, fail clearly — every error names what failed and what to check
- No secrets in output — key paths are OK, key contents never

## Test Server

```
Host: 46.101.213.82
User: root
Key:  ~/.ssh/id_rsa
```

## Detailed Docs

- [Go conventions](docs/GO.md) — code style, error handling, testify usage
- [Testing strategy](docs/TESTING.md) — TDD workflow, test tiers, integration tags
- [Architecture](docs/ARCHITECTURE.md) — module boundaries, data flow, stage pipeline
- [Output rules](docs/OUTPUT.md) — user-facing messages, progress format, error format
