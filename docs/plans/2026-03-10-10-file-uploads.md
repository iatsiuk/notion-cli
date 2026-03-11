# 10 - File Uploads: create, get, delete, send, complete

## References
- OpenAPI spec: `docs/openapi.json` (endpoints: /v1/file_uploads/*)
- Live docs: https://developers.notion.com/llms.txt

## Overview
- Implement `notion-cli file create` - initiate a file upload
- Implement `notion-cli file get <file_upload_id>` - get upload status
- Implement `notion-cli file delete <file_upload_id>` - delete upload
- Implement `notion-cli file send <file_upload_id> <file_path>` - upload file content
- Implement `notion-cli file complete <file_upload_id>` - mark upload as complete

## Context
- Depends on plans 01 (foundation), 02 (API client), 03 (output)
- File upload is a multi-step process: create -> send (multipart) -> complete
- Send endpoint uses multipart/form-data
- File uploads can be referenced in page/block content

## Development Approach
- **Testing approach**: TDD (tests first, then implementation)
- **CRITICAL: every task MUST include tests written BEFORE implementation code**
- **CRITICAL: all tests must pass before starting next task**

## Implementation Steps

### Task 1: File upload types and create API method
- [x] write tests in `internal/api/files_test.go`:
  - CreateFileUpload initiates upload
  - returns file upload object with ID
  - error handling
- [x] define FileUpload struct in `internal/api/files.go` matching openapi.json schema
- [x] implement CreateFileUpload method on Client (POST /v1/file_uploads)
- [x] run tests - must pass before next task

### Task 2: Get and delete file upload API methods
- [x] write tests for GetFileUpload and DeleteFileUpload:
  - get returns upload status
  - delete removes upload
  - error handling
- [x] implement GetFileUpload (GET /v1/file_uploads/{id}) and DeleteFileUpload (DELETE /v1/file_uploads/{id})
- [x] run tests - must pass before next task

### Task 3: Send file content API method
- [x] write tests for SendFileContent:
  - sends file as multipart/form-data
  - handles large files
  - error handling
- [x] implement SendFileContent method on Client (POST /v1/file_uploads/{id}/send)
- [x] run tests - must pass before next task

### Task 4: Complete file upload API method
- [x] write tests for CompleteFileUpload:
  - marks upload as complete
  - error handling
- [x] implement CompleteFileUpload method on Client (POST /v1/file_uploads/{id}/complete)
- [x] run tests - must pass before next task

### Task 5: File upload CLI commands
- [x] write tests for all file subcommands in `internal/cli/file_test.go`
- [x] implement `file create` command
- [x] implement `file get` command
- [x] implement `file delete` command
- [x] implement `file send` command (reads file from disk)
- [x] implement `file complete` command
- [x] run tests - must pass before next task

### Task 6: Convenience `file upload` command (all-in-one)
- [ ] write tests for `file upload <file_path>`:
  - creates upload, sends content, completes - all in one
  - outputs final file upload object
  - error handling at each step
- [ ] implement `file upload` command that chains create -> send -> complete
- [ ] run tests - must pass before next task

### Task 7: Verify acceptance criteria
- [ ] verify all file upload commands work
- [ ] verify multipart upload works
- [ ] verify all-in-one upload command
- [ ] run full test suite with `make test`
- [ ] run linter with `make build` - all issues must be fixed

## Technical Details
- FileUpload struct: id, object, status, filename, content_type, size
- POST /v1/file_uploads - body: { filename, content_type }
- GET /v1/file_uploads/{id}
- DELETE /v1/file_uploads/{id}
- POST /v1/file_uploads/{id}/send - multipart/form-data with file content
- POST /v1/file_uploads/{id}/complete - marks as done

## Post-Completion
- File uploads can be referenced in page/block creation
