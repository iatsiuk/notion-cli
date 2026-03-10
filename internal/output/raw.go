package output

import (
	"encoding/json"
	"fmt"
	"io"
)

type rawFormatter struct{}

// NewRawFormatter returns a Formatter that outputs strings unquoted,
// and other types as compact JSON.
func NewRawFormatter() Formatter {
	return &rawFormatter{}
}

func (f *rawFormatter) Format(w io.Writer, data any) error {
	if s, ok := data.(string); ok {
		_, err := fmt.Fprintln(w, s)
		return err
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("raw format: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}
