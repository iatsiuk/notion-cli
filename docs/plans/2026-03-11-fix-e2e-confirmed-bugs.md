# Fix E2E Confirmed Bugs

## Overview
Fix 3 confirmed bugs from e2e testing (`e2e-errors.md`):
1. **Table format** -- `unsupported type` for all API struct types (medium)
2. **File upload complete** -- explicit complete call fails on auto-completed single-part uploads (high)
3. **Block append TTY hang** -- stdin read blocks indefinitely on interactive terminal (medium)

## Context
- Files involved:
  - `internal/output/table.go` (lines 49-64 -- `toRows()` only handles `map[string]any`)
  - `internal/cli/file.go` (lines 179-216 -- `runFileUpload()` unconditionally calls complete)
  - `internal/api/files.go` (lines 128-139 -- `CompleteFileUpload()` no status check)
  - `internal/cli/block.go` (line 121 -- `io.ReadAll(stdin)` without TTY check)
  - `internal/cli/user.go` (lines 17-28 -- existing `isTerminal()` helper, output-only)
- Patterns: `httptest.NewServer()` for API mocking, table-driven tests, `t.Parallel()`
- Reference:
  - OpenAPI spec: `docs/openapi.json`
  - Live docs and spec: `https://developers.notion.com/llms.txt`

## Development Approach
- **Testing approach**: TDD (tests first)
- Complete each task fully before moving to the next
- Make small, focused changes
- **CRITICAL: every task MUST include new/updated tests** for code changes in that task
  - write tests BEFORE implementation code (TDD)
  - tests cover both success and error scenarios
- **CRITICAL: all tests must pass before starting next task**
- **CRITICAL: update this plan file when scope changes during implementation**
- Run `make build` after each change (runs linter automatically)

## Testing Strategy
- **Unit tests**: required for every task, written FIRST (TDD)
- Test patterns: table-driven, `t.Parallel()`, stdlib `testing` only
- Mocking: `httptest.NewServer()` for API, `strings.NewReader()` for stdin
- Run: `go test -race ./...`

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix
- Update plan if implementation deviates from original scope

## Implementation Steps

### Task 1: Table format -- write failing tests for struct types
- [x] add test case `TestTableFormat_APIUser` -- pass `*api.User` struct, expect formatted table output (not error)
- [x] add test case `TestTableFormat_APIUserSlice` -- pass `[]api.User`, expect multi-row table
- [x] add test case `TestTableFormat_APIDatabase` -- pass `*api.Database`, expect formatted table
- [x] add test case `TestTableFormat_NilAndEmpty` -- pass nil and empty slice, expect no error
- [x] run tests -- confirm they FAIL with "unsupported type" error

### Task 2: Table format -- fix toRows() to handle arbitrary structs
- [x] in `toRows()` (`internal/output/table.go:49-64`), add conversion path: marshal struct to JSON, then unmarshal to `map[string]any` (reuse json round-trip pattern from jsonl.go reflect approach)
- [x] handle both single struct (pointer) and slice of structs
- [x] run tests from Task 1 -- must pass
- [x] run `make build` -- linter must pass

### Task 3: File upload -- write failing test for auto-completed upload
- [x] add test `TestRunFileUpload_SkipsCompleteWhenAlreadyUploaded` -- mock server returns status="uploaded" after send, expect no complete call and success output
- [x] add test `TestRunFileUpload_CompletesWhenPending` -- mock server returns status="pending" after send, complete returns status="uploaded", expect success
- [x] add API-level test `TestSendFileContent_ReturnsStatus` -- verify `SendFileContent` returns updated `FileUpload` with status field
- [x] run tests -- confirm the skip-complete test FAILS (current code always calls complete)

### Task 4: File upload -- fix runFileUpload to check status before complete
- [x] modify `SendFileContent()` (`internal/api/files.go`) to return `*FileUpload` (parse response body) instead of just error, so caller can check status
- [x] in `runFileUpload()` (`internal/cli/file.go:200-209`), after send: check returned status, skip complete if status is already "uploaded"
- [x] run tests from Task 3 -- must pass
- [x] run `make build` -- linter must pass

### Task 5: Block append -- write failing test for TTY stdin detection
- [x] add `isInputTerminal(r io.Reader) bool` helper test in `internal/cli/block_test.go` -- test with `*os.File` (char device) returns true, `*strings.Reader` returns false
- [x] add test `TestRunBlockAppend_ErrorsOnTTYStdin` -- pass `*os.File` pointing to /dev/tty (or mock), expect immediate error "provide --children flag or pipe JSON array via stdin"
- [x] run tests -- confirm TTY test FAILS (current code hangs or reads empty)

### Task 6: Block append -- add TTY check before stdin read
- [ ] add `isInputTerminal(r io.Reader) bool` helper in `internal/cli/block.go` (similar to existing `isTerminal` but for `io.Reader`)
- [ ] in `runBlockAppend()` (`internal/cli/block.go:119-129`), before `io.ReadAll(stdin)`: check `isInputTerminal(stdin)`, return error immediately if true
- [ ] run tests from Task 5 -- must pass
- [ ] run `make build` -- linter must pass

### Task 7: Verify acceptance criteria
- [ ] verify table format works for all API types used in commands (User, Database, Page, Block, FileUpload, Comment, Search results)
- [ ] verify file upload succeeds end-to-end (single-part auto-complete scenario)
- [ ] verify block append returns error on TTY stdin without --children
- [ ] run full test suite: `go test -race ./...`
- [ ] run linter: `golangci-lint run`
- [ ] all e2e-errors.md bugs addressed

## Technical Details

### Table format fix
Current `toRows()` type switch only handles `[]json.RawMessage`, `[]any`, `map[string]any`. Fix: add `default` branch that uses `encoding/json` round-trip (marshal struct -> unmarshal to `map[string]any` or `[]map[string]any`). Use `reflect` to detect slice vs single value (pattern from `internal/output/jsonl.go`).

### File upload fix
Current flow: create -> send (returns error only) -> complete (fails if auto-completed). Fix: `SendFileContent` should return `*FileUpload` with status. If `status == "uploaded"`, skip complete call. Reference: Notion File Upload API -- single-part uploads may auto-complete after send.

### Block append TTY fix
Current flow: if `--children` not set, unconditionally `io.ReadAll(stdin)`. Fix: check if stdin is a terminal (`os.ModeCharDevice`), return descriptive error. Reuse pattern from existing `isTerminal()` in `user.go` but adapted for `io.Reader`.

## Post-Completion

**Manual verification**:
- Test `./notion-cli user me -f table` with real API token
- Test `./notion-cli file upload <file>` with real file
- Test `./notion-cli block append <id>` in interactive terminal (should error immediately)
- Re-run e2e test suite to confirm fixes
