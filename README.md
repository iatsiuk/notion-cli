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
| `json`  | Pretty-printed JSON             |
| `jsonl` | One JSON object per line        |
| `raw`   | Strings unquoted, other values as compact JSON |
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

Get the full page content rendered as Markdown. Always outputs plain Markdown text; `--format` is not applicable.

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
    --children string   Child blocks as non-empty JSON array
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

## Comment Commands

### comment list

List comments on a block or page.

```
notion-cli comment list [flags]

Flags:
    --block string   Block or page ID to list comments for (required)
```

```
notion-cli comment list --block abc123
```

### comment create

Create a comment on a page or reply in an existing discussion thread.

```
notion-cli comment create [flags]

Flags:
    --page string         Page ID to comment on
    --discussion string   Discussion ID to reply in
    --text string         Comment text content (required)
```

`--page` and `--discussion` are mutually exclusive; one is required.

```
notion-cli comment create --page abc123 --text "This looks good"
notion-cli comment create --discussion def456 --text "Agreed"
```

### comment get

Get a comment by ID.

```
notion-cli comment get <comment_id>
```

```
notion-cli comment get abc123
```

## User Commands

### user me

Get the currently authenticated user (the bot or user associated with the token).

```
notion-cli user me
```

```
notion-cli user me
```

### user list

List all users in the workspace.

```
notion-cli user list
```

```
notion-cli user list
notion-cli user list --format table
```

### user get

Get a workspace user by ID.

```
notion-cli user get <user_id>
```

```
notion-cli user get abc123
```

## File Upload Commands

### file create

Initiate a new file upload. Returns a file upload object with an ID used in subsequent steps.

```
notion-cli file create [flags]

Flags:
    --filename string          Filename for the upload
    --content-type string      MIME content type
    --mode string              Upload mode (e.g. multi for multipart)
    --number-of-parts int      Number of parts for multipart upload
```

```
notion-cli file create --filename report.pdf --content-type application/pdf
notion-cli file create --filename video.mp4 --mode multi --number-of-parts 3
```

### file get

Get a file upload object by ID.

```
notion-cli file get <file_upload_id>
```

```
notion-cli file get abc123
```

### file send

Upload file content to an existing file upload.

```
notion-cli file send <file_upload_id> <file_path> [flags]

Flags:
    --content-type string   MIME content type (auto-detected from extension if omitted)
    --part int              Part number for multipart upload (0 = single part)
```

```
notion-cli file send abc123 ./report.pdf
notion-cli file send abc123 ./chunk1.bin --part 1
```

### file complete

Mark a multipart file upload as complete after all parts have been sent.

```
notion-cli file complete <file_upload_id>
```

```
notion-cli file complete abc123
```

### file upload

Upload a file in one step: creates the upload, sends the content, and completes it automatically.

```
notion-cli file upload <file_path>
```

```
notion-cli file upload ./report.pdf
notion-cli file upload ./image.png
```

## Data Source Commands

### datasource list

List accessible Notion data sources.

```
notion-cli datasource list
```

```
notion-cli datasource list
notion-cli datasource list --format table
```

### datasource get

Get a Notion data source by ID.

```
notion-cli datasource get <data_source_id>
```

```
notion-cli datasource get abc123
```

### datasource create

Create a new Notion data source.

```
notion-cli datasource create [flags]

Flags:
    --parent string       Parent database: database_id:id (required)
    --title string        Data source title (plain text)
    --properties string   Properties as JSON object (required)
```

```
notion-cli datasource create --parent database_id:abc123 --title "Sales Data" --properties '{"Name":{"title":{}}}'
```

### datasource update

Update an existing Notion data source.

```
notion-cli datasource update <data_source_id> [flags]

Flags:
    --title string        New data source title (plain text)
    --properties string   Properties as JSON object
```

```
notion-cli datasource update abc123 --title "Updated Sales Data"
notion-cli datasource update abc123 --properties '{"Status":{"select":{}}}'
```

### datasource query

Query a Notion data source.

```
notion-cli datasource query <data_source_id> [flags]

Flags:
    --filter string   Filter as JSON object
```

```
notion-cli datasource query abc123
notion-cli datasource query abc123 --filter '{"property":"Status","select":{"equals":"Active"}}'
```

### datasource templates

Get templates for a Notion data source.

```
notion-cli datasource templates <data_source_id>
```

```
notion-cli datasource templates abc123
```

## OAuth Commands

Manage OAuth tokens. These commands use HTTP Basic authentication (client ID and secret) instead of a bearer token.

### oauth token

Exchange an authorization code for an access token.

```
notion-cli oauth token [flags]

Flags:
    --code string            Authorization code from OAuth redirect (required)
    --client-id string       OAuth client ID (required)
    --client-secret string   OAuth client secret (required)
    --redirect-uri string    Redirect URI used in authorization request
```

```
notion-cli oauth token --client-id abc123 --client-secret secret456 --code auth_code_here
notion-cli oauth token --client-id abc123 --client-secret secret456 --code auth_code_here --redirect-uri https://example.com/callback
```

### oauth introspect

Introspect an access token to retrieve its metadata.

```
notion-cli oauth introspect [flags]

Flags:
    --token string           Access token to introspect (required)
    --client-id string       OAuth client ID (required)
    --client-secret string   OAuth client secret (required)
```

```
notion-cli oauth introspect --client-id abc123 --client-secret secret456 --token ntn_token_here
```

### oauth revoke

Revoke an access token.

```
notion-cli oauth revoke [flags]

Flags:
    --token string           Access token to revoke (required)
    --client-id string       OAuth client ID (required)
    --client-secret string   OAuth client secret (required)
```

```
notion-cli oauth revoke --client-id abc123 --client-secret secret456 --token ntn_token_here
```

## Status

Check connectivity to the Notion API. Sends a request to verify that the configured token is valid and the API is reachable.

```
notion-cli status
```

Prints `ok` on success. Returns a non-zero exit code on error.


## Pipe-friendly Workflows

notion-cli outputs JSON by default on TTY and JSONL (one JSON object per line) when piped, making it easy to compose with jq and other Unix tools.

Extract all page titles from a database query (replace `Name` with your title property name):

```sh
notion-cli db query <database_id> | jq -r '.properties.Name.title[].plain_text'
```

Get IDs of all pages in a database whose Status is "Done":

```sh
notion-cli db query <database_id> \
  --filter '{"property":"Status","select":{"equals":"Done"}}' \
  | jq -r '.id'
```

Archive all pages returned by a search and print their titles:

```sh
notion-cli search "meeting notes" --type page | jq -r '.id' | while read id; do
  title=$(notion-cli page get "$id" | jq -r '[.properties[] | select(.type == "title")] | .[0].title[0].plain_text // "untitled"')
  notion-cli page update "$id" --archive
  echo "archived: $title"
done
```

List all workspace users and extract name/email pairs as TSV:

```sh
notion-cli user list | jq -r '[.name, .person.email // ""] | @tsv'
```
