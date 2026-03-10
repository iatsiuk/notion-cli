# 03 - Output Formatting: JSON, JSONL, Raw, Table

## References
- OpenAPI spec: `docs/openapi.json`
- Live docs: https://developers.notion.com/llms.txt

## Overview
- Implement output formatters: json, jsonl, raw, table
- Auto-detect format based on TTY (json) vs pipe (jsonl)
- Support --format flag override
- Formatters used by all command handlers

## Context
- Depends on plan 01 (config with format setting)
- README defines: json (pretty), jsonl (compact per line), raw (strings unquoted), table (ASCII aligned)
- Auto-detect: json on TTY, jsonl when piped

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- **CRITICAL: every task MUST include tests written BEFORE implementation code**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Formatter interface and JSON formatter
- [ ] write tests in `internal/output/json_test.go`:
  - pretty-prints JSON with indentation
  - handles arrays, objects, nested structures
  - handles empty/null values
- [ ] define Formatter interface in `internal/output/formatter.go`
- [ ] implement JSON formatter in `internal/output/json.go`
- [ ] run tests - must pass before next task

### Task 2: JSONL formatter
- [ ] write tests in `internal/output/jsonl_test.go`:
  - one compact JSON per line
  - handles arrays (each element on separate line)
  - handles single objects
- [ ] implement JSONL formatter in `internal/output/jsonl.go`
- [ ] run tests - must pass before next task

### Task 3: Raw formatter
- [ ] write tests in `internal/output/raw_test.go`:
  - strings output unquoted
  - numbers/booleans as compact JSON
  - objects/arrays as compact JSON
- [ ] implement raw formatter in `internal/output/raw.go`
- [ ] run tests - must pass before next task

### Task 4: Table formatter
- [ ] write tests in `internal/output/table_test.go`:
  - aligned columns with headers
  - handles variable-width content
  - handles missing/null fields
  - handles empty results
- [ ] implement table formatter in `internal/output/table.go`
- [ ] run tests - must pass before next task

### Task 5: Format auto-detection and factory
- [ ] write tests in `internal/output/factory_test.go`:
  - returns JSON formatter for TTY
  - returns JSONL formatter for pipe
  - returns correct formatter for explicit format flag
  - error for unknown format
- [ ] implement formatter factory function
- [ ] run tests - must pass before next task

### Task 6: Verify acceptance criteria
- [ ] verify all four formats produce correct output
- [ ] verify auto-detection works
- [ ] run full test suite with `make test`
- [ ] run linter with `make build` - all issues must be fixed

## Technical Details
- Formatter interface: `Format(w io.Writer, data any) error`
- JSON: `json.MarshalIndent` with 2-space indent
- JSONL: `json.Marshal` per item, newline separated
- Raw: strings unquoted via type assertion, rest as compact JSON
- Table: calculate column widths from headers + data, pad with spaces

## Post-Completion
- All command plans will use these formatters for output
