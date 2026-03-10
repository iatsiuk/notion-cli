package output_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"notion-cli/internal/output"
)

func TestNew_explicitFormats(t *testing.T) {
	t.Parallel()

	cases := []struct {
		format  string
		wantErr bool
	}{
		{"json", false},
		{"jsonl", false},
		{"raw", false},
		{"table", false},
		{"unknown", true},
		{"auto-detect", false}, // empty format falls back to json when not a tty
	}

	for _, tc := range cases {
		format := tc.format
		name := format
		if format == "auto-detect" {
			format = ""
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			f, err := output.New(format, false)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !errors.Is(err, output.ErrUnknownFormat) {
					t.Fatalf("expected ErrUnknownFormat, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f == nil {
				t.Fatal("expected non-nil formatter")
			}
		})
	}
}

func TestNew_autoDetect(t *testing.T) {
	t.Parallel()

	data := map[string]any{"key": "value"}

	// isTTY=true => JSON formatter (pretty-printed with indentation)
	f, err := output.New("", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var buf bytes.Buffer
	if err := f.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	if !strings.Contains(buf.String(), "\n  ") {
		t.Errorf("isTTY=true: expected indented JSON, got %q", buf.String())
	}

	// isTTY=false => JSONL formatter (compact, no indentation)
	f2, err := output.New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	buf.Reset()
	if err := f2.Format(&buf, data); err != nil {
		t.Fatalf("Format() error = %v", err)
	}
	out := buf.String()
	if strings.Contains(out, "\n  ") {
		t.Errorf("isTTY=false: expected compact JSONL, got %q", out)
	}
	if strings.Count(out, "\n") != 1 {
		t.Errorf("isTTY=false: expected single line output, got %q", out)
	}
}
