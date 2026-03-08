# ship

Build here. Run there.

If the source code and Docker cache live on your machine, but deployment happens on a remote host, there is an annoying gap: the server needs fresh images, but building there is slow, fragile, or not allowed.

`ship` closes that gap for Docker Compose projects. It builds images locally, transfers them to the remote host over SSH, restores their original tags, and runs your deploy command.

The problem it solves is simple:

- local builds are faster because the code and cache are already there
- remote hosts should run containers, not act like build machines
- copying source code around just to rebuild is unnecessary

`ship` keeps the build where it belongs and gets the result onto the host that needs to run it.

It is especially useful when a full registry setup feels like too much:

- running and securing a private registry is extra operational work
- paying for a hosted registry can be hard to justify for one app or one server
- pushing to Docker Hub just to redeploy a personal project is often the wrong level of ceremony

If SSH already exists, `ship` gives you a shorter path: build locally, transfer over the tunnel, deploy.

## How It Works

One `ship` run does this:

1. builds images from your Compose files
2. discovers which services actually produce images
3. tags those images as `localhost:5001/...`
4. starts a local registry on port `5001`
5. pushes the images into that registry
6. opens an SSH reverse tunnel so the remote host can reach it
7. pulls the images remotely, restores the original tags, and runs your command

## What It Does Not Do

`ship` does not:

- copy source code to the remote host
- sync Compose files to the remote host
- replace your deploy logic
- manage Docker on the remote host for you

The remote host still needs the project checkout if your deploy command expects it.

## When It Fits

Use `ship` when:

- the remote host already has the Compose project checked out
- images should be built locally or in CI, not on the server
- SSH access exists but a full registry setup feels like overkill
- you want to avoid registry credentials, registry operations, and another paid service for a small deployment

## Requirements

Local machine:

- Go `1.25.0+` if building from source
- Docker with Compose V2
- `ssh`
- the Compose file(s) you want to build from
- an SSH private key that can reach the remote host

Remote host:

- SSH access for the target user
- Docker installed and running
- the Compose project already present if your command expects it
- port `5001` available for the reverse tunnel

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
./ship \
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
- should restart services with the freshly transferred images

```bash
./ship \
  --docker-compose compose.yml,compose.prod.yml \
  --user deploy \
  --host staging.example.com \
  --key ~/.ssh/id_ed25519 \
  --command "cd /srv/app && docker compose -f compose.yml -f compose.prod.yml up -d"
```

If the remote host also needs to pull external images referenced by the Compose files, include that in your command:

```bash
./ship \
  --docker-compose compose.yml,compose.prod.yml \
  --user deploy \
  --host staging.example.com \
  --key ~/.ssh/id_ed25519 \
  --command "cd /srv/app && docker compose -f compose.yml -f compose.prod.yml pull && docker compose -f compose.yml -f compose.prod.yml up -d"
```

## Preflight Checks

Before the first stage starts, `ship` fails fast if any of these are missing or broken:

- Docker
- Docker Compose V2
- `ssh`
- the SSH key file
- every Compose file passed to `--docker-compose`
- SSH connectivity to `--user@--host`

If preflight passes, stage progress is printed in a consistent `[N/7]` format and the remote command output is passed through directly.

## Constraints

- only services with a Compose `build` key are transferred
- the transfer registry is fixed at `localhost:5001`
- the deploy path depends on SSH reverse port forwarding
- `ship` runs exactly the remote command you provide

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
