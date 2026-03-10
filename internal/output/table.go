package output

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type tableFormatter struct{}

// NewTableFormatter returns a Formatter that outputs data as an ASCII-aligned table.
func NewTableFormatter() Formatter {
	return &tableFormatter{}
}

func (f *tableFormatter) Format(w io.Writer, data any) error {
	rows, err := toRows(data)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	headers := collectHeaders(rows)
	widths := computeWidths(headers, rows)

	if err := writeRow(w, headers, widths); err != nil {
		return err
	}
	for _, row := range rows {
		vals := make([]string, len(headers))
		for i, h := range headers {
			vals[i] = cellValue(row[h])
		}
		if err := writeRow(w, vals, widths); err != nil {
			return err
		}
	}
	return nil
}

// toRows normalizes data into a slice of map[string]any rows.
func toRows(data any) ([]map[string]any, error) {
	if data == nil {
		return nil, nil
	}
	switch v := data.(type) {
	case []any:
		rows := make([]map[string]any, 0, len(v))
		for _, item := range v {
			m, ok := item.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("table format: expected object, got %T", item)
			}
			rows = append(rows, m)
		}
		return rows, nil
	case map[string]any:
		return []map[string]any{v}, nil
	default:
		return nil, fmt.Errorf("table format: unsupported type %T", data)
	}
}

// collectHeaders returns sorted column names from all rows.
func collectHeaders(rows []map[string]any) []string {
	seen := make(map[string]bool)
	for _, row := range rows {
		for k := range row {
			seen[k] = true
		}
	}
	headers := make([]string, 0, len(seen))
	for k := range seen {
		headers = append(headers, k)
	}
	sort.Strings(headers)
	return headers
}

// computeWidths returns max column widths for headers and all rows.
func computeWidths(headers []string, rows []map[string]any) []int {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, h := range headers {
			if w := len(cellValue(row[h])); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

// cellValue converts a cell value to its string representation.
func cellValue(v any) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}

// writeRow writes a single row with columns padded to widths.
func writeRow(w io.Writer, cols []string, widths []int) error {
	var sb strings.Builder
	for i, col := range cols {
		sb.WriteString(col)
		pad := widths[i] - len(col)
		if pad > 0 {
			sb.WriteString(strings.Repeat(" ", pad))
		}
		if i < len(cols)-1 {
			sb.WriteString("  ")
		}
	}
	sb.WriteByte('\n')
	_, err := fmt.Fprint(w, sb.String())
	return err
}
