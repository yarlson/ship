# Docker Utilities and Image Handling

## Overview

The `docker/` package provides CLI wrappers around Docker operations required by stages 1-2. All functions shell out to the `docker` CLI, no Go SDKs.

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
