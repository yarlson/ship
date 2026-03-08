# Context Map — Documentation Index

## Core Files

- `summary.md` — Project overview: What, Architecture, Core Flow, System State, Capabilities, Tech Stack
- `terminology.md` — Term definitions (Config, stage, local registry, etc.)
- `practices.md` — Conventions and invariants (code quality, testing, CLI, secrets)

## Domain Documentation

### CLI Domain

- `cli/parsing.md` — Flag parsing, validation, error handling

### Workflow Domain

- `workflow/stages.md` — 7-stage pipeline definition and flow

### Docker Domain

- `docker/image-handling.md` — Docker CLI operations and image metadata parsing

### Stage Domain

- `stage/implementation.md` — Stage 1-7 implementations, data flow, and integration patterns

### SSH Domain

- `ssh/tunnel-management.md` — SSH tunnel lifecycle, remote command execution, process management

### Progress Domain

- `progress/output.md` — Progress reporting format and testability

---

## Related Project Documentation

- `CLAUDE.md` — Project principles (DRY, KISS, SOLID, TDD, CLI-based, fail-fast)
- `docs/ARCHITECTURE.md` — Module boundaries, data flow, design decisions
- `docs/GO.md` — Go conventions, error handling, testing patterns
- `docs/TESTING.md` — Testing strategy and test tiers
- `docs/OUTPUT.md` — User-facing message and error format rules
