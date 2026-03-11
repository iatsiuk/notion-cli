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
