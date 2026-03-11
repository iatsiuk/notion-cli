# README Restructure

## Overview
- Rewrite README.md following the linear-cli README structure as reference
- Replace single "Quick Usage" code block with dedicated sections per command group
- Add missing sections: From source installation, Pipe-friendly Workflows
- Each command gets its own section with flags in code blocks and usage examples

## Context (from discovery)
- Files involved: `README.md`
- Reference: `/Volumes/Data/Devzone/linear-cli/README.md`
- Commands: status, user, page, db, block, comment, search, file, datasource, oauth
- Global flags: --token/-t, --format/-f, --quiet, --verbose
- No shell completion in codebase (skip that section)
- No config file support (authentication via env var / flag only)

## Development Approach
- **Testing approach**: N/A (documentation only)
- Single file change -- README.md
- Use actual command flags from cobra definitions (verified via codebase exploration)
- Keep all text in English (per CLAUDE.md)

## Progress Tracking
- Mark completed items with `[x]` immediately when done
- Add newly discovered tasks with + prefix
- Document issues/blockers with ! prefix

## Implementation Steps

### Task 1: Write header and installation sections
- [x] write project title and one-line description
- [x] write Homebrew installation section
- [x] write Binary releases section (link to GitHub Releases)
- [x] write From source section (Go 1.23+, git clone, make build)

### Task 2: Write authentication and configuration sections
- [ ] write Authentication section (NOTION_TOKEN env var, --token flag, precedence)
- [ ] write Global Flags table (--token, --format, --quiet, --verbose)
- [ ] write Output Formats section (auto-detection, json/jsonl/raw/table)
- [ ] write Exit Codes table

### Task 3: Write Search command section
- [ ] write `search` command with flags (--type, --sort) and examples

### Task 4: Write Page Commands section
- [ ] write `page get` with flags and example
- [ ] write `page create` with flags (--parent, --properties) and example
- [ ] write `page update` with flags (--properties, --archive/--unarchive) and example
- [ ] write `page property` with example
- [ ] write `page move` with flags (--parent) and example
- [ ] write `page markdown` with example

### Task 5: Write Database Commands section
- [ ] write `db get` with example
- [ ] write `db list` with example
- [ ] write `db create` with flags (--parent, --title, --properties) and example
- [ ] write `db update` with flags (--title, --description, --properties) and example
- [ ] write `db query` with flags (--filter, --sort) and examples

### Task 6: Write Block Commands section
- [ ] write `block get` with example
- [ ] write `block update` with flags (--data, --archive/--unarchive) and example
- [ ] write `block children` with example
- [ ] write `block append` with flags (--children) and example
- [ ] write `block delete` with example

### Task 7: Write Comment Commands section
- [ ] write `comment list` with flags (--block) and example
- [ ] write `comment create` with flags (--page/--discussion, --text) and examples
- [ ] write `comment get` with example

### Task 8: Write User Commands section
- [ ] write `user me` with example
- [ ] write `user list` with example
- [ ] write `user get` with example

### Task 9: Write File Upload Commands section
- [ ] write `file create` with flags and example
- [ ] write `file get` with example
- [ ] write `file send` with flags (--content-type, --part) and example
- [ ] write `file complete` with example
- [ ] write `file upload` (one-step) with example

### Task 10: Write Data Source Commands section
- [ ] write `datasource list` with example
- [ ] write `datasource get` with example
- [ ] write `datasource create` with flags and example
- [ ] write `datasource update` with flags and example
- [ ] write `datasource query` with flags and example
- [ ] write `datasource templates` with example

### Task 11: Write OAuth Commands section
- [ ] write `oauth token` with flags and example
- [ ] write `oauth introspect` with flags and example
- [ ] write `oauth revoke` with flags and example

### Task 12: Write Status command section
- [ ] write `status` command description and example

### Task 13: Write Pipe-friendly Workflows section
- [ ] write 3-4 pipeline examples using jq with notion-cli commands

### Task 14: Final review
- [ ] verify all commands from codebase are documented
- [ ] verify all flags match cobra definitions
- [ ] verify consistent formatting across all sections
- [ ] run `make build` to ensure no accidental code changes

## Technical Details
- Structure order: Installation -> Authentication -> Global Flags -> Output Formats -> Exit Codes -> Commands (search, page, db, block, comment, user, file, datasource, oauth, status) -> Pipe-friendly Workflows
- Each command section: H2 for command group, H3 for subcommands
- Flags in code blocks (same format as linear-cli)
- Examples in code blocks after each subcommand

## Post-Completion
- Review rendered markdown on GitHub for formatting issues
