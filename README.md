# notion-cli

Notion command-line client. Query and manage pages, databases, and blocks -- all from the terminal.

## Installation

### Homebrew

```bash
brew install iatsiuk/tap/notion-cli
```

### From releases

Download the binary from the [GitHub Releases](https://github.com/iatsiuk/notion-cli/releases) page.

## Quick Usage

```bash
# search pages
notion-cli search "Meeting notes"

# get a page
notion-cli page get <page-id>

# create a page
notion-cli page create --parent database_id:<db-id> --properties '{"Name":{"title":[{"text":{"content":"My page"}}]}}'

# update a page
notion-cli page update <page-id> --properties '{"Name":{"title":[{"text":{"content":"New title"}}]}}'

# archive / unarchive a page
notion-cli page update <page-id> --archive
notion-cli page update <page-id> --unarchive

# get a page property value
notion-cli page property <page-id> <property-id>

# move a page to a new parent
notion-cli page move <page-id> --parent page_id:<parent-page-id>

# get page content as markdown (raw text, ignores --format)
notion-cli page markdown <page-id>

# list databases
notion-cli db list

# query a database
notion-cli db query <database-id>

# server status
notion-cli status

# get current user
notion-cli user me

# list workspace users
notion-cli user list

# get a user by ID
notion-cli user get <user-id>
```

## Global Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--token` | `-t` | | Notion API token |
| `--format` | `-f` | *(auto)* | Output: json, jsonl, raw, table |
| `--quiet` | | false | Suppress non-data stderr output |
| `--verbose` | | false | Show request info and timing |

## Output Formats

Format is auto-detected: `json` (pretty-printed) on TTY, `jsonl` (one JSON per line) when piped. Override with `-f`:

- **json** -- pretty-printed JSON
- **jsonl** -- one compact JSON document per line
- **raw** -- strings unquoted, other values as compact JSON
- **table** -- aligned ASCII table

## Environment Variables

| Variable | Overrides |
|----------|-----------|
| `NOTION_TOKEN` | `--token` |

CLI flags always take precedence over environment variables.

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Connection error |
| 2 | API error |
| 3 | Authentication error |
