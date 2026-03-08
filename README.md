# ship

`ship` deploys Docker Compose images to a remote host without building on that host.

It builds your images locally, pushes them through a temporary local registry exposed over an SSH reverse tunnel, restores the original image tags on the remote machine, and then runs your deploy command.

This is useful when:

- builds should happen on the machine that already has source code and build cache
- the remote host should only pull images and run `docker compose`
- the remote host already has the Compose project checked out, but not the freshly built images

## What `ship` actually does

`ship` does not copy source code or Compose files to the remote host.

It does this in one run:

1. builds images from your local Compose file set
2. tags them as `localhost:5001/...`
3. starts a local Docker registry on port `5001`
4. pushes the images into that registry
5. opens an SSH reverse tunnel so the remote host can reach that registry
6. pulls the images on the remote host and restores their original tags
7. runs your remote deploy command

## What must already exist

On the local machine:

- Go `1.25.0+` if building from source
- Docker with Compose V2
- `ssh`
- the Compose file(s) you want to build from
- an SSH private key that can reach the remote host

On the remote host:

- `ssh` access for the target user
- Docker installed and running
- the Compose project already present if your command expects it
- a free port `5001` on the remote side for the reverse tunnel

## Install

Build from source:

```bash
go build -o ship .
```

Show help:

```bash
./ship --help
```

## Usage

```bash
ship \
  --docker-compose <compose.yml[,override.yml]> \
  --user <ssh-user> \
  --host <host> \
  --key <private-key-path> \
  --command "<remote-command>"
```

Required flags:

| Flag               | Meaning                                                     |
| ------------------ | ----------------------------------------------------------- |
| `--docker-compose` | One or more local Compose files, comma-separated            |
| `--user`           | SSH user on the remote host                                 |
| `--host`           | Remote host                                                 |
| `--key`            | Path to the SSH private key                                 |
| `--command`        | Command to run on the remote host after images are restored |

## Example

Local machine:

- has `compose.yml` and `compose.prod.yml`
- builds the images

Remote host:

- already has the same Compose project checked out
- should restart services with the newly shipped images

```bash
./ship \
  --docker-compose compose.yml,compose.prod.yml \
  --user deploy \
  --host staging.example.com \
  --key ~/.ssh/id_ed25519 \
  --command "cd /srv/app && docker compose -f compose.yml -f compose.prod.yml up -d"
```

If the remote host needs to pull external images referenced by the Compose file, do that in the command you pass:

```bash
./ship \
  --docker-compose compose.yml,compose.prod.yml \
  --user deploy \
  --host staging.example.com \
  --key ~/.ssh/id_ed25519 \
  --command "cd /srv/app && docker compose -f compose.yml -f compose.prod.yml pull && docker compose -f compose.yml -f compose.prod.yml up -d"
```

## What `ship` checks before it starts

Before running any stage, `ship` fails fast if any of these are missing or broken:

- Docker
- Docker Compose V2
- `ssh`
- the SSH key file
- every Compose file passed to `--docker-compose`
- SSH connectivity to `--user@--host`

If preflight passes, stage progress is printed in a consistent `[N/7]` format and the final remote command output is passed through directly.

## Constraints

- `ship` only transfers images built from services that have a Compose `build` key.
- The local registry port is fixed at `5001`.
- The transfer path depends on SSH reverse port forwarding to the remote host.
- `ship` restores original image tags on the remote host, then runs exactly the command you provide.

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

E2E tests against the remote host:

```bash
export SHIP_E2E_USER=deploy
export SHIP_E2E_HOST=staging.example.com
export SHIP_E2E_KEY=~/.ssh/id_ed25519
go test -race -count=1 -v -timeout=120s -tags=e2e ./...
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
- [Context docs](docs/context/)
