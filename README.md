# ship

Build, transfer, and deploy Docker Compose images to a remote host in a single command.

`ship` builds your Docker Compose images locally, transfers them to a remote host via SSH tunnel and a local registry, and executes a deployment command—all in one invocation.

- **Local build** — Builds images defined in your Docker Compose file(s)
- **Image transfer** — Pushes images through a local registry (port 5001) via SSH reverse tunnel
- **Remote deployment** — Executes your deployment command on the remote host

## Prerequisites

- Go 1.25.0 or later
- Docker (with Docker Compose V2)
- SSH access to the remote host with a private key

Before running the pipeline, `ship` validates all prerequisites:

- Docker is installed and accessible
- Docker Compose V2 plugin is available
- SSH is installed
- SSH key file exists and is readable
- Compose file(s) exist and are readable
- SSH connectivity to the remote host works

If any check fails, `ship` exits immediately with a clear error message before any stages run.

## Install

```bash
go install ship@latest
```

Or build from source:

```bash
go build -o ship .
```

## Quickstart

```bash
ship \
  --docker-compose docker-compose.yml \
  --user deploy \
  --host 10.0.0.5 \
  --key ~/.ssh/id_ed25519 \
  --command "docker compose up -d"
```

## Usage

```
ship [flags]
```

### Required Flags

| Flag                      | Description                                           |
| ------------------------- | ----------------------------------------------------- |
| `--docker-compose <path>` | Path to compose file(s), comma-separated for multiple |
| `--user <user>`           | SSH user on the remote host                           |
| `--host <host>`           | Remote host address                                   |
| `--key <path>`            | Path to SSH private key file                          |
| `--command <cmd>`         | Command to execute on the remote host after transfer  |

### Examples

Single compose file:

```bash
ship --docker-compose docker-compose.yml --user deploy --host 10.0.0.5 --key ~/.ssh/id_ed25519 --command "docker compose up -d"
```

Multiple compose files:

```bash
ship --docker-compose compose.yml,compose.prod.yml --user root --host staging.example.com --key ./key.pem --command "docker compose pull && docker compose up -d"
```

## Configuration

| Variable | Required | Description                                |
| -------- | -------- | ------------------------------------------ |
| `HOME`   | Yes      | Used for SSH key path expansion (implicit) |

## Workflow Stages

`ship` executes a 7-stage workflow:

| Stage | Description                                  | Status      |
| ----- | -------------------------------------------- | ----------- |
| 1     | Build Docker Compose images                  | Implemented |
| 2     | Re-tag images with `localhost:5001/` prefix  | Implemented |
| 3     | Ensure local registry container on port 5001 | Implemented |
| 4     | Push images to local registry                | Implemented |
| 5     | Establish SSH reverse tunnel                 | Implemented |
| 6     | Pull and restore images on remote host       | Implemented |
| 7     | Execute remote command via SSH               | Implemented |

## Troubleshooting

### Docker or Docker Compose not found

**Symptom**: `Error: Docker is not installed or not in PATH` or `Error: docker compose (V2) is required`

**Fix**:

- Install [Docker Desktop](https://www.docker.com/products/docker-desktop) or Docker Engine
- Ensure Docker Compose V2 is installed: `docker compose version`
- If using Docker Engine, install the [Compose plugin](https://docs.docker.com/compose/install/linux/#install-using-the-repository)

### SSH key file not found or unreadable

**Symptom**: `Error: SSH key file not found: /path/to/key — verify the --key path`

**Fix**:

- Verify the `--key` path is correct
- Use `~` for home directory expansion (e.g., `~/.ssh/id_ed25519`)
- Check file permissions: `ls -la ~/.ssh/id_ed25519`

### SSH connection failed

**Symptom**: `Error: SSH connection failed — verify --host and --key`

**Fix**:

- Verify the remote `--host` is reachable and correct
- Verify the `--key` matches a private key on your system
- Check SSH access: `ssh -i ~/.ssh/id_ed25519 user@host echo ok`
- Ensure the remote user (via `--user`) has SSH access configured

### Compose file not found or unreadable

**Symptom**: `Error: Compose file not found: /path/to/compose.yml` or `Error: Cannot read compose file: /path/to/compose.yml — check file permissions`

**Fix**:

- Verify the `--docker-compose` path(s) are correct and exist
- For relative paths, ensure they're relative to your current working directory
- Check file permissions: `ls -la compose.yml`
- Use absolute paths if relative paths cause issues
- For multiple files, ensure all files in the comma-separated list exist

### Port 5001 is required

**Symptom**: Cannot use a custom registry port.

**Cause**: Port 5001 is hardcoded in the implementation.

**Workaround**: Ensure port 5001 is available or modify the source code.

## Development

### Build

```bash
go build -o ship .
```

### Test

Run unit tests:

```bash
go test -race -count=1 -v -timeout=120s ./...
```

Run integration tests (requires Docker):

```bash
go test -race -count=1 -v -timeout=120s -tags=integration ./...
```

### Lint

```bash
golangci-lint run --fix ./...
```

### Format

```bash
bunx prettier --write "**/*.md"
```

## Project Structure

```
ship/
├── main.go          # Entry point — parse flags, invoke workflow
├── cli/             # Flag parsing and validation
├── workflow/        # 7-stage orchestrator
├── stage/           # Individual workflow stages
├── docker/          # Docker CLI wrappers
├── progress/        # Stage progress printer
└── testlock/        # Test synchronization utilities
```

## Documentation

- [Architecture](docs/ARCHITECTURE.md) — Module map, boundaries, data flow, design decisions
- [Testing](docs/TESTING.md) — TDD workflow, test tiers, Testify usage
- [Go Conventions](docs/GO.md) — Style, error handling, import order
- [Output Rules](docs/OUTPUT.md) — User-facing messages, progress format, error format

## Contributing

Not documented. Check the repository for contribution guidelines.

## License

Not specified.
