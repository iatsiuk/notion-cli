package output_test

import (
	"bytes"
	"testing"

	"notion-cli/internal/output"
)

func TestJSONLFormatter(t *testing.T) {
	t.Parallel()

	f := output.NewJSONLFormatter()

	tests := []struct {
		name  string
		input any
		want  string
	}{
		{
			name:  "array of objects - each on separate line",
			input: []any{map[string]any{"id": 1}, map[string]any{"id": 2}},
			want:  "{\"id\":1}\n{\"id\":2}\n",
		},
		{
			name:  "single object",
			input: map[string]any{"key": "value"},
			want:  "{\"key\":\"value\"}\n",
		},
		{
			name:  "empty array",
			input: []any{},
			want:  "",
		},
		{
			name:  "array of strings",
			input: []any{"foo", "bar"},
			want:  "\"foo\"\n\"bar\"\n",
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
