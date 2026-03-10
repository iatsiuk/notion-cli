package output_test

import (
	"bytes"
	"testing"

	"notion-cli/internal/output"
)

func TestTableFormatter(t *testing.T) {
	t.Parallel()

	f := output.NewTableFormatter()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "empty array",
			input: []any{},
			want:  "",
		},
		{
			name:  "nil value",
			input: nil,
			want:  "",
		},
		{
			name:  "array of empty objects",
			input: []any{map[string]any{}, map[string]any{}},
			want:  "",
		},
		{
			name: "utf8 multibyte column name and value",
			input: map[string]any{
				"nom": "café",
			},
			want: "nom \ncafé\n",
		},
		{
			name: "single object",
			input: map[string]any{
				"name": "Alice",
				"age":  float64(30),
			},
			// columns sorted alphabetically: age, name
			want: "age  name \n30   Alice\n",
		},
		{
			name: "array with aligned columns",
			input: []any{
				map[string]any{"name": "Alice", "city": "New York"},
				map[string]any{"name": "Bob", "city": "LA"},
			},
			// columns sorted: city, name
			want: "city      name \nNew York  Alice\nLA        Bob  \n",
		},
		{
			name: "handles missing fields",
			input: []any{
				map[string]any{"name": "Alice", "age": float64(30)},
				map[string]any{"name": "Bob"},
			},
			// columns sorted: age, name
			want: "age  name \n30   Alice\n     Bob  \n",
		},
		{
			name: "handles null fields",
			input: []any{
				map[string]any{"name": "Alice", "extra": nil},
				map[string]any{"name": "Bob", "extra": nil},
			},
			// columns sorted: extra, name
			want: "extra  name \n       Alice\n       Bob  \n",
		},
		{
			name: "handles variable-width content",
			input: []any{
				map[string]any{"col": "short"},
				map[string]any{"col": "a very long value here"},
			},
			want: "col                   \nshort                 \na very long value here\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			if err := f.Format(&buf, tt.input); err != nil {
				t.Fatalf("Format() error = %v", err)
			}
			if got := buf.String(); got != tt.want {
				t.Errorf("Format() =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestTableFormatter_errors(t *testing.T) {
	t.Parallel()

	f := output.NewTableFormatter()

	errCases := []struct {
		name  string
		input any
	}{
		{
			name:  "array with non-object element",
			input: []any{"not an object"},
		},
		{
			name:  "unsupported scalar type",
			input: 42,
		},
	}

	for _, tt := range errCases {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var buf bytes.Buffer
			if err := f.Format(&buf, tt.input); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
