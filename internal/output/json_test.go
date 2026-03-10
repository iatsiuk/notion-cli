package output_test

import (
	"bytes"
	"testing"

	"notion-cli/internal/output"
)

func TestJSONFormatter(t *testing.T) {
	t.Parallel()

	f := output.NewJSONFormatter()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "simple object",
			input: map[string]any{"key": "value"},
			want:  "{\n  \"key\": \"value\"\n}\n",
		},
		{
			name:  "array of objects",
			input: []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
			want:  "[\n  {\n    \"id\": 1\n  },\n  {\n    \"id\": 2\n  }\n]\n",
		},
		{
			name:  "nested structure",
			input: map[string]any{"outer": map[string]any{"inner": "val"}},
			want:  "{\n  \"outer\": {\n    \"inner\": \"val\"\n  }\n}\n",
		},
		{
			name:  "empty object",
			input: map[string]any{},
			want:  "{}\n",
		},
		{
			name:  "null value",
			input: nil,
			want:  "null\n",
		},
		{
			name:  "string value",
			input: "hello",
			want:  "\"hello\"\n",
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
