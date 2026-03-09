# ship

Transfer local Docker images to a remote host over SSH, without setting up a remote registry first.

`ship` is for solo developers and small teams who already build images on their own machines and want those exact images on a server. Instead of adding a separate build agent, pushing to a hosted registry, or shuffling tarballs around with `docker save` and `docker load`, `ship` lets the remote host pull the image straight from your machine over SSH.

The point is simple: build locally, transfer directly, keep the original tags, and keep Docker's normal layer-based transfer behavior.

## Why `ship` exists

For a lot of small setups, the developer machine is already the fastest and cheapest build machine available.

That leaves one awkward step: getting the image onto the remote host.

- `docker save` and `docker load` work, but they turn the transfer into a tarball workflow instead of a registry-style pull.
- A remote registry works too, but it adds setup, cost, and an extra hop: local machine -> registry -> remote host.
- Hosted build agents can be slower, more expensive, or both, especially when the image already exists locally.

`ship` takes a narrower path:

- use the image that already exists locally
- expose a temporary local registry on your machine
- tunnel that registry to the remote host over SSH
- pull the image on the remote host and restore the original tag

## What `ship` does

One `ship` run has one job: make sure one or more existing local Docker images end up on one remote host under their original tags.

The current workflow is:

1. tag each local image as `localhost:5001/<image>`
2. ensure a local `registry:2` container is running on port `5001`
3. push the transfer tags to that local registry
4. open an SSH reverse tunnel to the remote host
5. run remote `docker pull` and `docker tag` commands to restore the original image refs

Before any of that starts, `ship` checks:

- Docker is available locally
- `ssh` is available locally
- the SSH key exists if `-i` was provided
- each requested local image exists
- SSH connectivity to the target works

## What `ship` does not do

`ship` is intentionally narrow. It does not:

- build Docker images
- parse Docker Compose files
- copy source code
- run arbitrary remote commands
- restart containers after the transfer

If you need deployment orchestration, do that separately.

## Requirements

Local machine:

- Docker installed and responsive
- `ssh` installed
- the image or images already present locally
- port `5001` available for the local registry

Remote host:

- reachable over SSH
- Docker installed and running
- port `5001` available for the reverse tunnel endpoint

`-i` is optional. If you leave it out, `ship` falls back to normal SSH identity resolution.

## Install

### Homebrew

```bash
brew tap yarlson/tap
brew install ship
```

### GitHub Releases

Download the archive that matches your machine from the project's releases page, unpack it, and place the `ship` binary on your `PATH`.

Example:

```bash
VERSION=0.1.0
ARCHIVE=ship_${VERSION}_darwin_arm64.tar.gz

curl -LO https://github.com/yarlson/ship/releases/download/v${VERSION}/${ARCHIVE}
tar -xzf ${ARCHIVE}
chmod +x ship
sudo mv ship /usr/local/bin/ship
```

### Build from source

Requires Go `1.25`.

```bash
go build -o ship .
```

## Usage

```bash
ship [-i key] [-p port] user@host image[:tag] [image[:tag]...]
```

Examples:

```bash
ship root@10.0.0.1 app:latest
ship root@10.0.0.1 ship-test-api:latest traefik:v3
ship -i ~/.ssh/id_ed25519 deploy@staging.example.com app:latest
ship -i ~/.ssh/id_ed25519 -p 2222 deploy@staging.example.com ghcr.io/acme/app:dev redis:7
```

Help output:

```bash
ship --help
```

## Output

Progress is printed in a fixed `[N/5]` format:

```text
[1/5] Tagging images for transfer...
[1/5] Tag complete
[2/5] Starting local registry...
[2/5] Registry ready
[3/5] Pushing images to local registry...
[3/5] Push complete
[4/5] Establishing tunnel to staging.example.com...
[4/5] Tunnel established
[5/5] Pulling and restoring images on remote host...
[5/5] Pull and restore complete
Ship complete
  Host:     staging.example.com
  Images:   app:latest, redis:7
  Status:   Success
```

Errors are printed to stderr in `Error: ...` form and fail fast on the first problem.

## Development

Build:

```bash
go build -o ship .
```

Lint:

```bash
golangci-lint run --fix ./...
```

Unit tests:

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

## License

[MIT](LICENSE)
