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
