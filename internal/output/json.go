package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type jsonFormatter struct{}

// NewJSONFormatter returns a Formatter that pretty-prints JSON with 2-space indent.
func NewJSONFormatter() Formatter {
	return &jsonFormatter{}
}

func (f *jsonFormatter) Format(w io.Writer, data any) error {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("json format: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}
