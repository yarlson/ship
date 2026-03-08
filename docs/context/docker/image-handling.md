# Docker Utilities and Image Handling

## Overview

The `docker/` package provides CLI wrappers around Docker operations required by stages 1-4. All functions shell out to the `docker` CLI, no Go SDKs.

## Image Struct

```go
type Image struct {
    Name string
    Tag  string
}

func (i Image) Ref() string {
    return i.Name + ":" + i.Tag
}
```

**Invariant:** Name never contains a colon. Tag is always present (defaults to "latest").

## Core Operations

### ComposeBuild(composeFiles []string) error

Executes: `docker compose -f file1 -f file2 ... build`

- Connects stdout/stderr directly to parent process for passthrough output
- Returns error if build fails
- Called by Stage 1 to build all services with build keys

**Error handling:** Wraps exec error as `docker compose build: <err>`

### ComposeConfig(composeFiles []string) ([]Image, error)

Executes: `docker compose -f file1 -f file2 ... config --format json`

- Parses JSON output to extract images from services
- Only includes services with non-null "build" key
- Uses explicit image name or derives from compose project name + service name
- Returns list of Image structs

**Error handling:** Two failure points:

- Command execution: `docker compose config: <err>`
- JSON parsing: `parse compose config: <err>`

### TagImage(source, target string) error

Executes: `docker tag source target`

- Called by Stage 2 for each image in ImageMap
- Source: original image reference (e.g., "web:latest")
- Target: transfer tag (e.g., "localhost:5001/web:latest")

**Error handling:** Returns `docker tag source → target: <output>` with combined stdout/stderr

## Image Naming Rules

### parseImageRef(ref string) Image

Parses an image reference string into Name and Tag:

- Reference format: `[registry/]name[:tag]`
- Last colon only counts if it appears after the last slash (to handle registry prefixes)
- Default tag is "latest" if no tag present

Examples:

- `web:latest` → Image{Name: "web", Tag: "latest"}
- `ghcr.io/org/app:v1` → Image{Name: "ghcr.io/org/app", Tag: "v1"}
- `localhost:5001/web` → Image{Name: "localhost:5001/web", Tag: "latest"}
- `nginx` → Image{Name: "nginx", Tag: "latest"}

### TransferTag(img Image) string

Returns: `localhost:5001/` + img.Ref()

Examples:

- Image{Name: "web", Tag: "latest"} → "localhost:5001/web:latest"
- Image{Name: "api", Tag: "v2"} → "localhost:5001/api:v2"

## Docker Compose Config Parsing

The `docker compose config --format json` output is parsed to extract services:

```json
{
  "name": "my-project",
  "services": {
    "web": {
      "build": { ... },        // Non-null = service is built
      "image": "web:latest"    // Explicit image name
    },
    "db": {
      "image": "postgres:13"   // No build key = not built, skipped
    }
  }
}
```

**Extraction logic:**

1. Iterate through services
2. Skip services without "build" key (or with null build)
3. Extract image name from "image" field
4. If image name missing, derive as `<project-name>-<service-name>`
5. Parse image name into Image struct

**Result:** List of Image structs for all services that were built.

## Registry Operations (Stages 3-4)

All registry operations shell out to `docker` CLI or use raw TCP dialing for connection checks.

### CheckRegistryRunning() (bool, error)

Executes: `docker ps --filter ancestor=registry:2 --filter publish=5001 --format {{.ID}}`

- Checks if a `registry:2` container is running with port 5001 published
- Uses container filtering to find exact match
- Returns true if exactly one container found, false if none

**Error handling:** Returns `docker ps: <err>` if command fails.

### CheckPortConflict() (bool, error)

Performs two-stage port conflict detection:

**Stage 1 (Docker containers):**

- Executes: `docker ps --filter publish=5001 --format {{.ID}}`
- Returns true immediately if any container uses :5001
- Error handling: Returns `docker ps: <err>` if command fails

**Stage 2 (Non-Docker processes):**

- Attempts TCP dial to `localhost:5001` with 1-second timeout
- Returns true if connection succeeds (port occupied)
- Returns false if connection refused (port free) — treated as success, not error
- Used to detect processes running outside Docker

### StartRegistry() error

Executes: `docker run -d -p 5001:5000 registry:2`

**Process:**

1. Runs `registry:2` container in detached mode, maps host :5001 to container :5000
2. Waits for registry to accept TCP connections (up to 3 seconds)
3. Uses TCP dial with 500ms timeout, polls up to 30 times (3 second total wait)
4. Returns immediately on first successful connection

**Error handling:**

- Command execution: Returns `docker run registry: <output>`
- Timeout: Returns `registry started but not accepting connections on port 5001`
- Output includes docker run stderr if startup fails

### PushImage(imageRef string) error

Executes: `docker push imageRef`

- Pushes the specified image reference to registry (must be resolvable)
- Uses combined stdout/stderr for error reporting
- Called once per transfer-tagged image

**Error handling:** Returns `docker push: <output>` with full docker output on failure.

**Invariant:** ImageRef must be a valid transfer tag (e.g., `localhost:5001/web:latest`).
