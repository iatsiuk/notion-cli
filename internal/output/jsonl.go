package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
)

type jsonlFormatter struct{}

// NewJSONLFormatter returns a Formatter that writes one compact JSON per line.
// Arrays are expanded: each element on its own line.
func NewJSONLFormatter() Formatter {
	return &jsonlFormatter{}
}

func (f *jsonlFormatter) Format(w io.Writer, data any) error {
	if data == nil {
		_, err := fmt.Fprintln(w, "null")
		return err
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Slice {
		for i := range v.Len() {
			b, err := json.Marshal(v.Index(i).Interface())
			if err != nil {
				return fmt.Errorf("jsonl format: %w", err)
			}
			if _, err := fmt.Fprintf(w, "%s\n", b); err != nil {
				return err
			}
		}
		return nil
	}

	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("jsonl format: %w", err)
	}
	_, err = fmt.Fprintf(w, "%s\n", b)
	return err
}
