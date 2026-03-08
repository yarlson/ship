# Testing Strategy

## TDD Workflow

1. Write a failing CLI or workflow test for the transfer behavior.
2. Write the smallest failing unit or stage test underneath it.
3. Implement the minimum code to pass.
4. Refactor while green.

The outer contract is now small enough that most changes should start from the CLI or workflow layer.

## Test Tiers

| Tier        | Scope                                      | Tag           | Docker required |
| ----------- | ------------------------------------------ | ------------- | --------------- |
| Unit        | Parsing, formatting, validation            | none          | No              |
| Integration | Real Docker on the local machine           | `integration` | Yes             |
| E2E         | Full transfer path over SSH to a real host | `e2e`         | Yes + SSH       |

## Run Commands

```bash
# Unit only
go test -race -count=1 -v -timeout=120s ./...

# Unit + integration
go test -race -count=1 -v -timeout=120s -tags=integration ./...

# E2E against the real remote host
GOCACHE=/tmp/ship-gocache go test -race -count=1 -v -timeout=120s -tags=e2e ./...
```

Set E2E target configuration with environment variables:

```bash
export SHIP_E2E_USER=root
export SHIP_E2E_HOST=46.101.213.82
# Optional if SSH defaults are not enough:
export SHIP_E2E_KEY=~/.ssh/id_ed25519
```

## Integration Test Rules

- Gate with `//go:build integration`
- Use real Docker commands
- Clean up registry containers and transfer tags when needed

## E2E Test Rules

- Gate with `//go:build e2e`
- Require a real SSH host and a reachable Docker daemon
- `SHIP_E2E_KEY` is optional; tests may use the default SSH identity behavior
- Skip cleanly when Docker or SSH prerequisites are unavailable
