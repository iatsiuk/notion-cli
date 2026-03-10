package output

import (
	"errors"
	"fmt"
)

// ErrUnknownFormat is returned when an unrecognized format name is requested.
var ErrUnknownFormat = errors.New("unknown format")

// New returns a Formatter for the given format name.
// If format is empty, auto-detection is used: JSON for TTY, JSONL for pipes.
func New(format string, isTTY bool) (Formatter, error) {
	if format == "" || format == "auto" {
		if isTTY {
			return NewJSONFormatter(), nil
		}
		return NewJSONLFormatter(), nil
	}

	switch format {
	case "json":
		return NewJSONFormatter(), nil
	case "jsonl":
		return NewJSONLFormatter(), nil
	case "raw":
		return NewRawFormatter(), nil
	case "table":
		return NewTableFormatter(), nil
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnknownFormat, format)
	}
}
