package config

import (
	"errors"
	"os"
)

// Config holds the resolved CLI configuration.
type Config struct {
	Token   string
	Format  string
	Quiet   bool
	Verbose bool
}

// Load builds a Config from flag values and environment variables.
// tokenFlag and formatFlag are the values passed via CLI flags (empty if not set).
// NOTION_TOKEN env var is used as fallback when tokenFlag is empty.
// Format "auto" or "" resolves to "json" for TTY, "jsonl" for non-TTY.
func Load(tokenFlag, formatFlag string, quiet, verbose bool) (*Config, error) {
	token := os.Getenv("NOTION_TOKEN")
	if tokenFlag != "" {
		token = tokenFlag
	}
	if token == "" {
		return nil, errors.New("no token: set NOTION_TOKEN env var or --token flag")
	}

	isTTY := isTerminal()

	return &Config{
		Token:   token,
		Format:  resolveFormat(formatFlag, isTTY),
		Quiet:   quiet,
		Verbose: verbose,
	}, nil
}

// resolveFormat returns the output format based on flag value and terminal detection.
func resolveFormat(flagVal string, isTTY bool) string {
	if flagVal != "" && flagVal != "auto" {
		return flagVal
	}
	if isTTY {
		return "json"
	}
	return "jsonl"
}

// isTerminal reports whether os.Stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
