package output_test

import (
	"bytes"
	"testing"

	"notion-cli/internal/output"
)

func TestRawFormatter(t *testing.T) {
	t.Parallel()

	f := output.NewRawFormatter()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "string is unquoted",
			input: "hello world",
			want:  "hello world\n",
		},
		{
			name:  "empty string",
			input: "",
			want:  "\n",
		},
		{
			name:  "number as compact JSON",
			input: 42,
			want:  "42\n",
		},
		{
			name:  "float as compact JSON",
			input: 3.14,
			want:  "3.14\n",
		},
		{
			name:  "boolean true as compact JSON",
			input: true,
			want:  "true\n",
		},
		{
			name:  "boolean false as compact JSON",
			input: false,
			want:  "false\n",
		},
		{
			name:  "object as compact JSON",
			input: map[string]any{"key": "value"},
			want:  "{\"key\":\"value\"}\n",
		},
		{
			name:  "array as compact JSON",
			input: []any{"a", "b"},
			want:  "[\"a\",\"b\"]\n",
		},
		{
			name:  "null value",
			input: nil,
			want:  "null\n",
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
