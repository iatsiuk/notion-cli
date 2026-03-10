# Notion-CLI Project Instructions

## Package Structure

- `cmd/notion-cli` - CLI entry point
- `internal/api` - Notion API HTTP client, error types, pagination helper
- `internal/cli` - CLI commands and exit code handling
- `internal/output` - output formatters (json, jsonl, raw, table) and formatter factory; use `output.New(format, isTTY)` - caller determines TTY; default: json on TTY, jsonl on pipe
- `internal/` - internal packages

## Code Style

### Imports

Group imports in order, separated by blank lines:
1. Standard library
2. External packages
3. Local packages (`notion-cli/...`)

```go
import (
    "context"
    "fmt"

    "github.com/spf13/cobra"

    "notion-cli/internal/api"
)
```

### Naming

- Package names: short, lowercase, no underscores (`api`, `config`, `query`)
- Exported types: PascalCase (`Config`, `Client`, `Page`)
- Unexported: camelCase (`fetchPage`, `buildRequest`)
- Acronyms: consistent case (`URL`, `HTTP`, `API` or `url`, `http`, `api`)
- Receivers: short, 1-2 letters (`c` for `*Client`, `p` for `*Page`)
- Errors: `Err` prefix for sentinel errors (`ErrNotFound`)

### Functions

- Max 80 lines, 50 statements (enforced by `funlen` linter)
- Max cyclomatic complexity: 10 (enforced by `cyclop` linter)
- Max nesting depth: 5 (enforced by `nestif` linter)
- Early returns for error handling
- Group related functions together

### Error Handling

- Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- Check all errors (enforced by `errcheck` linter)
- Use `errors.Is`/`errors.As` for error comparison
- Sentinel errors as package-level variables

### Comments

- Only for non-obvious logic
- English, lowercase, brief
- No comments for self-explanatory code

### Structs

- JSON tags on all exported fields: `json:"field_name"`
- Use `omitempty` for optional fields
- Pointer types for optional values (`*float64`, `*int`)
- Group related fields together

### Variables

- Package-level constants in `const` block
- Related constants grouped together
- Unexported package variables with `var`

### Control Flow

- Use `range` with index for modifying slices
- Prefer `for i := range n` over `for i := 0; i < n; i++` (Go 1.22+)
- Use `switch` over long `if-else` chains

### Concurrency

- Use `context.Context` as first parameter
- Use `sync.Mutex` for simple locking
- Use `errgroup` for parallel operations

## Testing

- Use table-driven tests for multiple scenarios
- Use stdlib `testing` package only (no testify)
- Test error paths: timeouts, context cancellation
- Run with race detector: `go test -race ./...`
- Use `t.Parallel()` for independent tests
- Test files: `*_test.go` in same package
- Integration tests require Docker; run with `make test-integration` (uses `-tags integration`)

## Notion API

- OpenAPI spec: `docs/openapi.json`
- Live docs and spec: `https://developers.notion.com/llms.txt`

## Language

All documentation, comments, and text must be in English.

## Building

- Always build with `make build` (runs linter automatically)
- Direct `go build` skips linting - avoid it

## Linting and Formatting

- Run `golangci-lint run` before committing (executed automatically via `make build`)
- Fix formatting issues with `goimports -w <file>` or `gofmt -w <file>`
- Config: `.golangci.yml` defines enabled linters
- No trailing whitespace, proper import grouping (stdlib, external, local)
