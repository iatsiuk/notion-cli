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
# search pages and databases
notion-cli search "Meeting notes"

# filter by object type
notion-cli search "Meeting notes" --type page
notion-cli search "My DB" --type database

# sort by last edited time
notion-cli search "Meeting notes" --sort descending
notion-cli search "Meeting notes" --sort ascending

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

# get a database
notion-cli db get <database-id>

# list databases
notion-cli db list

# create a database
notion-cli db create --parent page_id:<page-id> --title "My DB" --properties '{}'
notion-cli db create --parent workspace --title "My DB"

# update a database
notion-cli db update <database-id> --title "New Title"
notion-cli db update <database-id> --description "New description"
notion-cli db update <database-id> --properties '{"Status":{"select":{}}}'

# query a database
notion-cli db query <database-id>
notion-cli db query <database-id> --filter '{"property":"Status","select":{"equals":"Done"}}'
notion-cli db query <database-id> --sort '[{"property":"Name","direction":"ascending"}]'

# server status
notion-cli status

# get current user
notion-cli user me

# list workspace users
notion-cli user list

# get a user by ID
notion-cli user get <user-id>

# get a block
notion-cli block get <block-id>

# update a block's type-specific content
notion-cli block update <block-id> --data '{"paragraph":{"rich_text":[{"text":{"content":"Hello"}}]}}'

# archive / unarchive a block
notion-cli block update <block-id> --archive
notion-cli block update <block-id> --unarchive

# list child blocks
notion-cli block children <block-id>

# append child blocks
notion-cli block append <block-id> --children '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"text":{"content":"New"}}]}}]'

# delete a block
notion-cli block delete <block-id>

# list comments on a block or page
notion-cli comment list --block <block-id>

# create a comment on a page
notion-cli comment create --page <page-id> --text "Hello"

# create a comment in a discussion thread
notion-cli comment create --discussion <discussion-id> --text "Reply"

# get a comment
notion-cli comment get <comment-id>
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
