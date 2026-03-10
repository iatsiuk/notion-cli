# 11 - Data Sources: list, create, get, update, query, templates

## References
- OpenAPI spec: `docs/openapi.json` (endpoints: /v1/data_sources/*)
- Live docs: https://developers.notion.com/llms.txt

## Overview
- Implement `notion-cli datasource list` - list data sources
- Implement `notion-cli datasource create` - create a data source
- Implement `notion-cli datasource get <id>` - get data source
- Implement `notion-cli datasource update <id>` - update data source
- Implement `notion-cli datasource query <id>` - query data source
- Implement `notion-cli datasource templates <id>` - get templates

## Context
- Depends on plans 01 (foundation), 02 (API client), 03 (output)
- Data sources are external data integrations
- Query endpoint returns structured data
- Templates endpoint returns available templates for a data source

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- **CRITICAL: every task MUST include tests written BEFORE implementation code**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: Data source types and list API method
- [ ] write tests in `internal/api/datasources_test.go`:
  - ListDataSources returns paginated results
  - data source JSON deserialization
  - error handling
- [ ] define DataSource struct in `internal/api/datasources.go` matching openapi.json schema
- [ ] implement ListDataSources method on Client (GET /v1/data_sources)
- [ ] run tests - must pass before next task

### Task 2: Create, get, update API methods
- [ ] write tests for CreateDataSource, GetDataSource, UpdateDataSource:
  - create returns new data source
  - get returns data source by ID
  - update modifies data source
  - error handling for each
- [ ] implement all three methods on Client
- [ ] run tests - must pass before next task

### Task 3: Query and templates API methods
- [ ] write tests for QueryDataSource and GetDataSourceTemplates:
  - query returns structured data
  - templates returns available templates
  - pagination handling
  - error handling
- [ ] implement QueryDataSource (POST /v1/data_sources/{id}/query)
- [ ] implement GetDataSourceTemplates (GET /v1/data_sources/{id}/templates)
- [ ] run tests - must pass before next task

### Task 4: Data source CLI commands
- [ ] write tests for all datasource subcommands in `internal/cli/datasource_test.go`
- [ ] implement `datasource list` command
- [ ] implement `datasource create` command
- [ ] implement `datasource get` command
- [ ] implement `datasource update` command
- [ ] implement `datasource query` command with --filter flag
- [ ] implement `datasource templates` command
- [ ] run tests - must pass before next task

### Task 5: Verify acceptance criteria
- [ ] verify all datasource commands work
- [ ] verify query with filters
- [ ] verify pagination in list and query
- [ ] run full test suite with `make test`
- [ ] run linter with `make build` - all issues must be fixed

## Technical Details
- DataSource struct: id, object, type, name, config
- GET /v1/data_sources - paginated list
- POST /v1/data_sources - create
- GET /v1/data_sources/{id} - get
- PATCH /v1/data_sources/{id} - update
- POST /v1/data_sources/{id}/query - query with filter
- GET /v1/data_sources/{id}/templates - get templates

## Post-Completion
- Data sources are standalone, no downstream dependencies
