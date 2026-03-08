# Terminology

**CLI Config** — Parsed command-line flags as a typed struct (`cli.Config`). Contains: compose files, SSH user/host/key, and deployment command.

**Local registry** — Docker registry running on `localhost:5001` used as an intermediate storage point to push images locally before transferring to remote.

**Reverse SSH tunnel** — Background SSH process that forwards the remote host's port to the local registry port, allowing remote pulls from localhost:5001.

**Stage** — One step in the 7-stage deployment pipeline (e.g., "Build images", "Push to local registry"). Each stage has a start and completion message.

**WorkflowState** — Shared mutable state passed through stages containing image metadata and tunnel process handle (future implementation).

**Compose file(s)** — Docker Compose YAML file(s) passed via `--docker-compose` flag. Can be comma-separated for multiple files.

**Image tagging** — Re-tagging Docker images from their original names (e.g., `web:latest`) to local registry format (e.g., `localhost:5001/web:latest`) for local registry push.

**ImageMap** — A map data structure built in Stage 1 that tracks the correspondence between original image references and their transfer tags. Key: original image ref; Value: `localhost:5001/` prefixed transfer tag. Passed through stages 2, 4 to support tagging and pushing operations.

**Transfer tag** — The `localhost:5001/` prefixed image tag used for local registry operations. Derived from the original image name and tag (e.g., `web:latest` → `localhost:5001/web:latest`).

**Compose config** — Output from `docker compose config --format json`. Used in Stage 1 to discover which services have build keys and extract their image names.

**Progress reporting** — Stage progress printed in `[N/7] message` format to help users track deployment progress.

**Fail fast** — Exit immediately on first error with a clear message identifying what failed and what to check.
