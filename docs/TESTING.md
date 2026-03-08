# Testing Strategy

## TDD Workflow (Outside-In)

1. Write a failing acceptance/integration test describing the desired behavior
2. Write a failing unit test for the smallest piece needed
3. Write minimal code to pass the unit test
4. Refactor while green
5. Repeat until the acceptance test passes

Start from the outermost layer (CLI, workflow) and work inward (stages, docker/ssh wrappers).

## Test Tiers

| Tier        | Scope                                       | Tag           | Docker required |
| ----------- | ------------------------------------------- | ------------- | --------------- |
| Unit        | Pure logic: parsing, formatting, validation | (none)        | No              |
| Integration | Real Docker and local filesystem checks     | `integration` | Yes             |
| E2E         | Full workflow against the real remote host  | `e2e`         | Yes + SSH       |

### Run Commands

```bash
# Unit only
go test -race -count=1 -v -timeout=120s ./...

# Unit + integration
go test -race -count=1 -v -timeout=120s -tags=integration ./...

# E2E against the real remote host
go test -race -count=1 -v -timeout=120s -tags=e2e ./...

# Full non-unit suite
go test -race -count=1 -v -timeout=120s -tags='integration e2e' ./...
```

Set E2E target configuration with environment variables:

```bash
export SHIP_E2E_USER=deploy
export SHIP_E2E_HOST=staging.example.com
export SHIP_E2E_KEY=~/.ssh/id_ed25519
```

## Integration Test Rules

- Gate with build tag: `//go:build integration`
- Use real Docker commands — no mocks for Docker
- Clean up containers/images created during tests

## E2E Test Rules

- Gate with build tag: `//go:build e2e`
- Configure the SSH test host via `SHIP_E2E_USER`, `SHIP_E2E_HOST`, and `SHIP_E2E_KEY`
- Expect Docker and SSH access to be available
- Keep these tests out of the default `integration` tag

## Testify Usage

- Use `require` for preconditions and setup assertions (fail immediately)
- Use `assert` for the actual test assertions when multiple checks make sense
- Prefer `require.NoError(t, err)` over `if err != nil { t.Fatal(err) }`
- Use `assert.Contains`, `assert.Equal`, `assert.ErrorContains` for clarity

## What Not to Test

- SSH tunnel reliability (infrastructure-dependent, E2E only)
- Docker daemon behavior (trust Docker; test our invocation and parsing)
- The `ssh` binary itself
