# notion-cli

Notion command-line client. Query and manage pages, databases, and blocks from the terminal.

## Installation

### Homebrew

```
brew install iatsiuk/tap/notion-cli
```

### Binary releases

Download pre-built binaries for Linux and macOS from
[GitHub Releases](https://github.com/iatsiuk/notion-cli/releases).

### From source

Requirements: Go 1.25+

```
git clone https://github.com/iatsiuk/notion-cli
cd notion-cli
make build
```

The binary is placed at `./notion-cli`. Move it to a directory in your PATH:

```
mv notion-cli /usr/local/bin/notion-cli
```

## Authentication

Set the `NOTION_TOKEN` environment variable to your Notion integration token:

```
export NOTION_TOKEN=secret_...
```

The `--token` flag overrides the environment variable for a single invocation:

```
notion-cli --token secret_... page get <page-id>
```

Token precedence: `--token` flag > `NOTION_TOKEN` env var.

## Global Flags

```
-t, --token string    Notion API token (overrides NOTION_TOKEN)
-f, --format string   Output format: auto|json|jsonl|raw|table (default "auto")
    --quiet           Suppress non-essential output
    --verbose         Enable verbose output
```

## Output Formats

The default format is `auto`: JSON for interactive terminals, JSON Lines for pipes.

| Format  | Description                     |
|---------|---------------------------------|
| `json`  | Pretty-printed JSON array       |
| `jsonl` | One JSON object per line        |
| `raw`   | Raw API response                |
| `table` | Human-readable table            |
| `auto`  | `json` on TTY, `jsonl` on pipe  |

## Exit Codes

| Code | Meaning                     |
|------|-----------------------------|
| 0    | Success                     |
| 1    | Connection or network error |
| 2    | API error                   |
| 3    | Authentication error        |

## Search

Search across all pages and databases in the workspace.

```
notion-cli search <query> [flags]

Flags:
    --type string   Filter by object type: page or database
    --sort string   Sort direction: ascending or descending
```

```
notion-cli search "project roadmap"
notion-cli search "meeting notes" --type page
notion-cli search "team wiki" --type database --sort descending
```

## Page Commands

### page get

Get a page by ID.

```
notion-cli page get <page_id>
```

```
notion-cli page get abc123
```

### page create

Create a new page under a database or page parent.

```
notion-cli page create [flags]

Flags:
    --parent string       Parent: type:id (e.g. database_id:abc or page_id:abc) (required)
    --properties string   Properties as JSON object (default "{}")
```

```
notion-cli page create --parent database_id:abc123 --properties '{"Name":{"title":[{"text":{"content":"New Page"}}]}}'
notion-cli page create --parent page_id:abc123
```

### page update

Update page properties or archive/unarchive a page.

```
notion-cli page update <page_id> [flags]

Flags:
    --properties string   Properties as JSON object (default "{}")
    --archive             Archive the page
    --unarchive           Unarchive the page
```

```
notion-cli page update abc123 --properties '{"Status":{"select":{"name":"Done"}}}'
notion-cli page update abc123 --archive
notion-cli page update abc123 --unarchive
```

### page property

Get a single page property by property ID.

```
notion-cli page property <page_id> <property_id>
```

```
notion-cli page property abc123 title
```

### page move

Move a page to a new parent page or data source.

```
notion-cli page move <page_id> [flags]

Flags:
    --parent string   New parent: type:id (e.g. page_id:abc or data_source_id:abc) (required)
```

```
notion-cli page move abc123 --parent page_id:def456
notion-cli page move abc123 --parent data_source_id:ghi789
```

### page markdown

Get the full page content rendered as Markdown.

```
notion-cli page markdown <page_id>
```

```
notion-cli page markdown abc123
notion-cli page markdown abc123 > page.md
```

## Database Commands

### db get

Get a database by ID.

```
notion-cli db get <database_id>
```

```
notion-cli db get abc123
```

### db list

List all databases accessible to the integration.

```
notion-cli db list
```

```
notion-cli db list
notion-cli db list --format table
```

### db create

Create a new database under a page or workspace.

```
notion-cli db create [flags]

Flags:
    --parent string       Parent: page_id:id or workspace (required)
    --title string        Database title (plain text)
    --properties string   Properties schema as JSON object
```

```
notion-cli db create --parent page_id:abc123 --title "My Tasks"
notion-cli db create --parent workspace --title "Projects" --properties '{"Name":{"title":{}},"Status":{"select":{}}}'
```

### db update

Update a database title, description, or properties schema.

```
notion-cli db update <database_id> [flags]

Flags:
    --title string        New database title (plain text)
    --description string  New database description (plain text)
    --properties string   Properties schema as JSON object
```

```
notion-cli db update abc123 --title "Updated Title"
notion-cli db update abc123 --description "Tracks all tasks" --properties '{"Priority":{"select":{}}}'
```

### db query

Query pages in a database with optional filters and sorts.

```
notion-cli db query <database_id> [flags]

Flags:
    --filter string   Filter as JSON object
    --sort string     Sorts as JSON array
```

```
notion-cli db query abc123
notion-cli db query abc123 --filter '{"property":"Status","select":{"equals":"Done"}}'
notion-cli db query abc123 --sort '[{"property":"Name","direction":"ascending"}]'
```

## Block Commands

### block get

Get a block by ID.

```
notion-cli block get <block_id>
```

```
notion-cli block get abc123
```

### block update

Update block content or archive/unarchive a block.

```
notion-cli block update <block_id> [flags]

Flags:
    --data string   Block type content as JSON object (default "{}")
    --archive       Archive the block
    --unarchive     Unarchive the block
```

```
notion-cli block update abc123 --data '{"paragraph":{"rich_text":[{"text":{"content":"Updated text"}}]}}'
notion-cli block update abc123 --archive
notion-cli block update abc123 --unarchive
```

### block children

List child blocks of a block.

```
notion-cli block children <block_id>
```

```
notion-cli block children abc123
```

### block append

Append child blocks to a block. Children can be provided via the `--children` flag or piped via stdin.

```
notion-cli block append <block_id> [flags]

Flags:
    --children string   Child blocks as JSON array (default "[]")
```

```
notion-cli block append abc123 --children '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"text":{"content":"New paragraph"}}]}}]'
echo '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"text":{"content":"New paragraph"}}]}}]' | notion-cli block append abc123
```

### block delete

Delete a block by ID.

```
notion-cli block delete <block_id>
```

```
notion-cli block delete abc123
```
