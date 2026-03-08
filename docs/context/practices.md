# Practices

## Product Boundaries

- Ship is an image transfer tool, not a deployment orchestrator.
- Keep the scope at one or more local images to one remote host.
- Do not add Docker Compose parsing or build orchestration back into the core path.
- Do not add arbitrary remote command execution as part of the transfer workflow.

## Code Conventions

- Use the standard library `flag` package for CLI parsing.
- Shell out to `docker` and `ssh`, not Go SDKs.
- Prefer direct error messages with one clear remediation hint.
- Keep stage functions narrow and explicit.

## Workflow Conventions

- Run preflight before any stage starts.
- Print `progress.StageStart()` before doing the stage work.
- Print `progress.StageComplete()` only after the stage succeeds.
- Keep tunnel cleanup in `workflow`, not inside the stage call sites.

## Testing Conventions

- Unit tests cover parsing, formatting, and small logic branches.
- Integration tests use real local Docker behavior.
- E2E tests use a real SSH host and real Docker image transfer.
- Capture output through `progress.Writer` instead of asserting on raw stdout where possible.

## Output Conventions

- Errors are lowercase Go error strings; `main.go` adds the `Error:` prefix.
- Success summaries show the original image refs, never the `localhost:5001/` transfer tags.
- No secrets in output. SSH key paths are acceptable; key contents are not.
