# Progress Reporting

## Format

Progress output follows a consistent `[N/7]` format to help users track deployment through all 7 stages.

**Stage start:**
```
[N/7] message...
```

**Stage completion:**
```
[N/7] message
```

(Note: Start messages include trailing ellipsis, completion messages do not.)

## Example Output

```
[1/7] Building images...
[1/7] Build complete
[2/7] Tagging images for transfer...
[2/7] Tag complete
[3/7] Starting local registry...
[3/7] Registry ready
[4/7] Pushing images to local registry...
[4/7] Push complete
[5/7] Establishing tunnel...
[5/7] Tunnel established
[6/7] Pulling and restoring images on remote host...
[6/7] Pull and restore complete
[7/7] Running remote command...
[7/7] Command complete
```

## Implementation

**File:** `progress/progress.go`

**API:**
- `StageStart(stage int, msg string)` — Print `[N/7] message...`
- `StageComplete(stage int, msg string)` — Print `[N/7] message`

**Testability:** `Writer` var allows tests to mock output destination instead of checking stdout.

```go
var Writer io.Writer = os.Stdout  // Can be replaced in tests

// Tests:
buf := &strings.Builder{}
progress.Writer = buf
// ... run code ...
// assert buf.String() matches expected output
```

## Invariants

- Total stages is always 7 (hardcoded constant)
- N ranges 1–7 in order
- No newlines or formatting beyond stage messages
