package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

// Config holds the resolved CLI configuration.
type Config struct {
	Token   string
	Format  string
	Quiet   bool
	Verbose bool
}

// ErrNoToken is returned when no API token is configured.
var ErrNoToken = errors.New("no token: set NOTION_TOKEN env var or --token flag")

// Load builds a Config from flag values and environment variables.
// tokenFlag and formatFlag are the values passed via CLI flags (empty if not set).
// NOTION_TOKEN env var is used as fallback when tokenFlag is empty.
// Format "auto" or "" resolves to "json" for TTY, "jsonl" for non-TTY.
func Load(tokenFlag, formatFlag string, quiet, verbose bool) (*Config, error) {
	token := strings.TrimSpace(os.Getenv("NOTION_TOKEN"))
	if tokenFlag != "" {
		token = strings.TrimSpace(tokenFlag)
	}
	if token == "" {
		return nil, ErrNoToken
	}

	isTTY := isTerminal()

	format, err := resolveFormat(formatFlag, isTTY)
	if err != nil {
		return nil, err
	}

	return &Config{
		Token:   token,
		Format:  format,
		Quiet:   quiet,
		Verbose: verbose,
	}, nil
}

var validFormats = map[string]bool{"json": true, "jsonl": true, "raw": true, "table": true}

// resolveFormat returns the output format based on flag value and terminal detection.
func resolveFormat(flagVal string, isTTY bool) (string, error) {
	if flagVal == "" || flagVal == "auto" {
		if isTTY {
			return "json", nil
		}
		return "jsonl", nil
	}
	if !validFormats[flagVal] {
		return "", fmt.Errorf("invalid format %q: must be one of json, jsonl, raw, table, auto", flagVal)
	}
	return flagVal, nil
}

// LoadNoToken loads config without requiring an API token.
// Used by commands that authenticate via other means (e.g., OAuth Basic auth).
func LoadNoToken(formatFlag string, quiet, verbose bool) (*Config, error) {
	isTTY := isTerminal()
	format, err := resolveFormat(formatFlag, isTTY)
	if err != nil {
		return nil, err
	}
	return &Config{Format: format, Quiet: quiet, Verbose: verbose}, nil
}

// isTerminal reports whether os.Stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
