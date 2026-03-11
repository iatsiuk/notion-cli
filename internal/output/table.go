package output

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"unicode/utf8"
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
	if len(headers) == 0 {
		return nil
	}
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
	case []json.RawMessage:
		return rawMessagesToRows(v)
	case []any:
		return anySliceToRows(v)
	case map[string]any:
		return []map[string]any{v}, nil
	default:
		return structToRows(data)
	}
}

// structToRows handles arbitrary struct/pointer/slice types via JSON round-trip.
func structToRows(data any) ([]map[string]any, error) {
	rv := reflect.ValueOf(data)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, nil
		}
		rv = rv.Elem()
	}
	switch rv.Kind() {
	case reflect.Slice:
		return jsonSliceToRows(data, rv.Len())
	case reflect.Struct:
		return jsonStructToRow(data)
	default:
		return nil, fmt.Errorf("table format: unsupported type %T", data)
	}
}

func jsonSliceToRows(data any, length int) ([]map[string]any, error) {
	if length == 0 {
		return nil, nil
	}
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("table format: marshal: %w", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(b, &rows); err != nil {
		return nil, fmt.Errorf("table format: decode slice: %w", err)
	}
	return rows, nil
}

func jsonStructToRow(data any) ([]map[string]any, error) {
	b, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("table format: marshal: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("table format: decode struct: %w", err)
	}
	return []map[string]any{m}, nil
}

func rawMessagesToRows(items []json.RawMessage) ([]map[string]any, error) {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		var m map[string]any
		if err := json.Unmarshal(item, &m); err != nil {
			return nil, fmt.Errorf("table format: decode item: %w", err)
		}
		rows = append(rows, m)
	}
	return rows, nil
}

func anySliceToRows(items []any) ([]map[string]any, error) {
	rows := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			rows = append(rows, map[string]any{})
			continue
		}
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("table format: expected object, got %T", item)
		}
		rows = append(rows, m)
	}
	return rows, nil
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
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i, h := range headers {
			if w := utf8.RuneCountInString(cellValue(row[h])); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

// cellValue converts a cell value to its string representation.
// Maps and slices are serialized as compact JSON.
func cellValue(v any) string {
	if v == nil {
		return ""
	}
	switch v.(type) {
	case map[string]any, []any:
		b, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(b)
	}
	return fmt.Sprintf("%v", v)
}

// writeRow writes a single row with columns padded to widths.
func writeRow(w io.Writer, cols []string, widths []int) error {
	var sb strings.Builder
	for i, col := range cols {
		sb.WriteString(col)
		pad := widths[i] - utf8.RuneCountInString(col)
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
