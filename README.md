# ship

Transfer a local Docker image to a remote host over SSH, without setting up a remote registry.

`ship` is for the annoying middle ground where:

- the image already exists on your machine
- the server needs that image
- building on the server is the wrong move
- Docker Hub or a private registry feels like extra ceremony

`ship` does one job: move a single image to a remote host through a local registry exposed over an SSH reverse tunnel, then restore the original tag on the remote side.

## What It Does

One `ship` run does this:

1. verifies the local image exists
2. tags it as `localhost:5001/<image>`
3. starts a local registry on port `5001`
4. pushes the transfer tag into that registry
5. opens an SSH reverse tunnel to the remote host
6. pulls the transfer tag on the remote host
7. restores the original image tag on the remote host

That is it.

`ship` does not:

- build images
- read Docker Compose files
- copy source code
- run deploy hooks or remote commands

If you want to restart containers after the transfer, do that separately.

## Usage

```bash
ship [-i key] [-p port] user@host image[:tag]
```

Examples:

```bash
ship root@10.0.0.1 app:latest
ship -i ~/.ssh/id_ed25519 deploy@staging.example.com app:latest
ship -i ~/.ssh/id_ed25519 -p 2222 deploy@staging.example.com ghcr.io/acme/app:dev
```

## Requirements

Local machine:

- Docker
- `ssh`
- the image already present locally

Remote host:

- SSH access
- Docker installed and running
- port `5001` available for the reverse tunnel

`-i` is optional. If omitted, `ship` uses the same default SSH identity behavior as `ssh`.

## Preflight Checks

Before stage 1 starts, `ship` verifies:

- Docker is available locally
- `ssh` is available locally
- the SSH key exists if `-i` was provided
- the local image exists
- SSH connectivity to the remote host works

If preflight passes, progress is printed in a consistent `[N/5]` format.

## Install

Build from source:

```bash
go build -o ship .
```

Show help:

```bash
./ship --help
```

## Development

Build:

```bash
go build -o ship .
```

Lint:

```bash
golangci-lint run --fix ./...
```

Test:

```bash
go test -race -count=1 -v -timeout=120s ./...
```

Integration tests:

```bash
go test -race -count=1 -v -timeout=120s -tags=integration ./...
```

E2E tests:

```bash
export SHIP_E2E_USER=root
export SHIP_E2E_HOST=46.101.213.82
# Optional if SSH defaults are not enough:
export SHIP_E2E_KEY=~/.ssh/id_ed25519
GOCACHE=/tmp/ship-gocache go test -race -count=1 -v -timeout=120s -tags=e2e ./...
```

Format docs:

```bash
bunx prettier --write "**/*.md"
```

## Docs

- [Architecture](docs/ARCHITECTURE.md)
- [Testing](docs/TESTING.md)
- [Go conventions](docs/GO.md)
- [Output rules](docs/OUTPUT.md)
