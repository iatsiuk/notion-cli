package output_test

import (
	"errors"
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
		{"", false}, // auto-detect falls back to json when not a tty
	}

	for _, tc := range cases {
		t.Run(tc.format, func(t *testing.T) {
			t.Parallel()
			f, err := output.New(tc.format, false)
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

	// isTTY=true => JSON formatter
	f, err := output.New("", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f == nil {
		t.Fatal("expected non-nil formatter")
	}

	// isTTY=false => JSONL formatter
	f2, err := output.New("", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f2 == nil {
		t.Fatal("expected non-nil formatter")
	}
}
