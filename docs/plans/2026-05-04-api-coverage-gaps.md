# notion-cli API coverage gaps: block append --after, page create --children

## Overview

Close three gaps in `notion-cli` against the Notion REST API surfaced in `issues.md` during a real working session:

1. `block append` does not expose the `after` field of `PATCH /v1/blocks/{block_id}/children`, so blocks can only be appended at the end of the parent.
2. `page create` does not expose the `children` field of `POST /v1/pages`, forcing two round-trips (create empty page + append) and leaving a half-created page on partial failure.
3. `block append --children` accepts only a bare JSON array `[...]`. When a user already has a full request body shaped as `{"children":[...], "after":"..."}` (or got one back from `curl`), they have to unwrap it via `jq`. The flag should also accept the object form.

These are flag-level gaps. The HTTP client is the only consumer of `internal/api`, so adjusting `AppendBlockChildren`'s signature is safe.

## Context (from discovery)

Files involved:
- `internal/api/blocks.go` - `AppendBlockChildren(ctx, blockID, children)` at lines 115-138; body assembled as `map[string]any{"children": children}`.
- `internal/api/blocks_test.go` - existing unit tests for append.
- `internal/cli/block.go` - `NewBlockAppendCmd` and `runBlockAppend` at lines 103-152; flag `--children` defined at line 115.
- `internal/cli/block_test.go` - existing CLI tests for `block append`.
- `internal/cli/page.go` - `NewPageCreateCmd` (lines 59-75) and `runPageCreate` (lines 255-283); flags `--parent` and `--properties` only.
- `internal/cli/page_test.go` - existing CLI tests for `page create`.
- `internal/api/pages.go` - `CreatePageRequest.Children []any` already declared with `omitempty` (line 48). No struct changes needed.

Patterns found:
- API methods use stdlib `net/http`; tests use `httptest.Server` to assert request bodies.
- CLI commands wired via cobra; tests run cobra commands end-to-end against `httptest.Server` set as the API base URL via env.
- Flag parsing: `cmd.Flags().Changed("flagname")` is used to detect explicit user input (already present in `block append` for `--children` to gate stdin reading).
- Style: stdlib `testing`, table-driven tests, no testify. `funlen` <= 80 lines, `cyclop` <= 10.

Dependencies identified: cobra (already in use), stdlib only for tests.

## Development Approach

- **Testing approach**: TDD-leaning per project rules - write/extend the table-driven tests that pin down new behavior alongside each implementation change in the same task.
- Complete each task fully before moving to the next.
- Make small, focused changes.
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task.
- **CRITICAL: all tests must pass before starting next task** - no exceptions.
- **CRITICAL: update this plan file when scope changes during implementation**.
- Run `make build` (includes golangci-lint) after each task.
- Maintain backward compatibility for users: existing array form of `--children` keeps working unchanged; existing absence of `--children` for `page create` keeps current behavior (no `children` in POST body).

## Testing Strategy

- **Unit tests**: required for every task (see Development Approach above).
- **API client tests**: use `httptest.Server` to assert exact request body shape (presence/absence of `after` and `children` fields).
- **CLI tests**: cobra command run end-to-end against `httptest.Server`, with assertions on request body and on validation errors before any HTTP call.
- Project has no Playwright/Cypress UI suite; integration tests in `internal/api/*_integration_test.go` (build tag `integration`) are out of scope here.

## Progress Tracking

- Mark completed items with `[x]` immediately when done.
- Add newly discovered tasks with the prefix-plus-arrow marker.
- Document blockers with the warning marker.
- Update plan if implementation deviates from original scope.

## What Goes Where

- **Implementation Steps** (`[ ]` checkboxes): code, tests, doc strings inside cobra flags.
- **Post-Completion** (no checkboxes): real-token smoke testing, README touch-ups, `issues.md` cleanup.

## Implementation Steps

### Task 1: Extend AppendBlockChildren API client signature with `after`

- [x] in `internal/api/blocks.go`, change `AppendBlockChildren(ctx, blockID, children []map[string]any) ([]Block, error)` to `AppendBlockChildren(ctx, blockID, children []map[string]any, after string) ([]Block, error)`
- [x] inside the function, set `body["after"] = after` only when `after != ""` (do not include the field when empty)
- [x] keep the existing `len(children) == 0` precondition error
- [x] update the existing call site in `internal/cli/block.go` to pass empty `""` so the project still compiles after this task (final wiring happens in Task 3)
- [x] in `internal/api/blocks_test.go`, add table-driven cases to the existing `TestAppendBlockChildren`-style test (or add a new one) that decode the captured request body and assert: case `after=""` -> body has no `after` key; case `after="block-xyz"` -> body has `"after":"block-xyz"`
- [x] run `go test -race ./internal/api/...` - must pass before next task

### Task 2: Add `--after` flag and JSON-object support to `block append`

- [x] in `internal/cli/block.go`, add a new persistent flag `cmd.Flags().StringVar(&afterFlag, "after", "", "Insert new children after this block ID (default: append to end)")` to `NewBlockAppendCmd`
- [x] update `--children` help text to: `"Child blocks as JSON array, or full request body as JSON object {\"children\":[...], \"after\":\"...\"}"`
- [x] extend `runBlockAppend` signature to accept `afterFlag` and `afterFlagSet bool` (`cmd.Flags().Changed("after")`)
- [x] introduce a small parser `parseAppendChildren(raw string) (children []map[string]any, afterFromJSON string, hadAfterKey bool, err error)`:
  - trim whitespace; reject empty input with the existing message
  - try `json.Unmarshal` into `map[string]any` first; if that succeeds AND the object contains the key `"children"`:
    - decode `children` as `[]map[string]any` (re-marshal+unmarshal the value)
    - if `"after"` key is present, capture it as a string; non-string -> error with field name
    - `hadAfterKey` reflects key presence (even if value is `""`)
    - reject any unknown top-level keys other than `children` and `after` to keep the contract tight
  - if the input is not a JSON object, fall back to the current behavior: unmarshal into `[]map[string]any`
- [x] in `runBlockAppend`, after parsing:
  - if `afterFlagSet && hadAfterKey` -> return error: `"conflicting after: provided via --after and inside --children JSON object"`
  - effective `after` = `afterFlag` when `afterFlagSet`, else `afterFromJSON`
  - keep the existing `len(children) == 0` -> error
  - call `client.AppendBlockChildren(ctx, blockID, children, after)`
- [x] in `internal/cli/block_test.go`, add table-driven cases:
  - `--after block-xyz` with array `--children` -> request body has both `children` and `"after":"block-xyz"`
  - `--children` object form with `after` inside -> request body has `after` from JSON
  - `--children` object form without `after` -> no `after` in body
  - `--children` object with `after` + `--after` flag -> validation error before any HTTP call
  - `--children` object with no `children` key -> validation error
  - `--children` object with unknown extra key (e.g. `parent`) -> validation error
- [x] run `go test -race ./internal/cli/...` - must pass before next task

### Task 3: Add `--children` flag to `page create`

- [x] in `internal/cli/page.go`, add `cmd.Flags().StringVar(&childrenFlag, "children", "[]", "Child blocks as JSON array (default: empty)")` to `NewPageCreateCmd`
- [x] thread `childrenFlag` into `runPageCreate`
- [x] in `runPageCreate`, parse `--children` strictly as a JSON array of objects (`[]map[string]any`):
  - empty array -> leave `req.Children` nil (omitempty drops it from POST body)
  - non-empty array -> assign to `req.Children` (convert to `[]any` to match the struct field type)
  - non-array (object, scalar, null) -> error `"--children: must be a JSON array"`
  - invalid JSON -> error wrapping the unmarshal error
- [x] in `internal/cli/page_test.go`, add table-driven cases:
  - happy path: `--children='[{...}]'` -> POST body contains `children` with the same shape
  - default `--children='[]'` -> POST body has no `children` key
  - flag not provided at all -> POST body has no `children` key
  - `--children='{"children":[]}'` -> validation error (object form rejected)
  - `--children='not-json'` -> validation error
- [x] run `go test -race ./internal/cli/...` - must pass before next task

### Task 4: Verify acceptance criteria

- [x] `make build` passes (golangci-lint clean)
- [x] `go test -race ./...` passes
- [x] `go run ./cmd/notion-cli block append --help` mentions both `--after` and the updated `--children` description
- [x] `go run ./cmd/notion-cli page create --help` mentions `--children`
- [x] grep test files to confirm new cases exist:
  - `internal/api/blocks_test.go` contains a check for the `after` field in request body
  - `internal/cli/block_test.go` contains a case for the `--after` + JSON `after` conflict
  - `internal/cli/page_test.go` contains a case for `--children` array path and for omission

## Technical Details

### API surface change

Before:
```go
func (c *Client) AppendBlockChildren(ctx context.Context, blockID string, children []map[string]any) ([]Block, error)
```

After:
```go
func (c *Client) AppendBlockChildren(ctx context.Context, blockID string, children []map[string]any, after string) ([]Block, error)
```

Body assembly:
```go
body := map[string]any{"children": children}
if after != "" {
    body["after"] = after
}
```

### `--children` parsing (block append)

Detection by JSON shape:
- Try unmarshal into `map[string]any`.
  - On success, require key `children`. Disallow keys outside `{children, after}`.
  - Decode `children` into `[]map[string]any` via re-marshal/unmarshal (works because nested values are preserved as `map[string]any`).
  - Capture `after` if the key exists; track presence separately from emptiness so users can distinguish "explicitly empty" from "absent" in error messages if needed.
- On failure, try unmarshal into `[]map[string]any` (legacy path).

Conflict rule: `--after` set AND `after` key present in object -> error, no HTTP call.

### `--children` parsing (page create)

Strict array. Object form is intentionally rejected to keep the UX consistent with the existing `--parent` / `--properties` split. Object form belongs to the API request body abstraction, which the CLI does not expose for `page create`.

### Help text updates

- `block append --after`: `"Insert new children after this block ID (default: append to end)"`
- `block append --children`: `"Child blocks as JSON array, or full request body as JSON object {\"children\":[...], \"after\":\"...\"}"`
- `page create --children`: `"Child blocks as JSON array (default: empty)"`

### Backward compatibility

- Array form of `--children` keeps working unchanged.
- Stdin path for `block append` keeps working: read raw bytes, pass through the same `parseAppendChildren` so stdin can also feed object form.
- Default value `--children="[]"` for `page create` is a no-op against the API.

### Git branch

`feature/notion-cli-api-gaps/block-after-page-children`

## Post-Completion

*Items requiring manual intervention or external systems - no checkboxes, informational only.*

**Manual verification (recommended, not automated):**
- Run a real `block append --after <block_id> --children '[...]'` against a sandbox Notion workspace; confirm the new blocks land after the specified anchor, not at the end.
- Run a real `page create --parent page_id:<id> --properties '{...}' --children '[...]'` and confirm the page is created with content in a single round-trip.

**External system updates:**
- README/docs updates if the project advertises CLI surface anywhere outside cobra help (none expected).
- `issues.md` cleanup is up to the operator (file is currently untracked; not part of this plan).

**Out-of-plan:**
- Release / Homebrew bump.
- Removing `issues.md`.
- Object form of `--children` for `page create`.
- `db create --children` (Notion API does not accept blocks at database creation).

## Link to original task

Local file `issues.md` in the repository root (untracked at the time of plan creation, dated 2026-05-04).
