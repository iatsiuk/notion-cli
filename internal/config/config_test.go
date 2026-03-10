package config

import (
	"testing"
)

func TestLoad_TokenFromEnv(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "env-token")

	cfg, err := Load("", "json", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Errorf("got token %q, want %q", cfg.Token, "env-token")
	}
}

func TestLoad_FlagOverridesEnv(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "env-token")

	cfg, err := Load("flag-token", "json", false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "flag-token" {
		t.Errorf("got token %q, want %q", cfg.Token, "flag-token")
	}
}

func TestLoad_NoToken(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "")

	_, err := Load("", "json", false, false)
	if err == nil {
		t.Error("expected error when no token provided")
	}
}

func TestLoad_QuietFlag(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "tok")

	cfg, err := Load("", "json", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Quiet {
		t.Error("expected Quiet=true")
	}
	if cfg.Verbose {
		t.Error("expected Verbose=false")
	}
}

func TestLoad_VerboseFlag(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "tok")

	cfg, err := Load("", "json", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Verbose {
		t.Error("expected Verbose=true")
	}
}

func TestResolveFormat_TTY(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flag string
		want string
	}{
		{"", "json"},
		{"auto", "json"},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			t.Parallel()
			got := resolveFormat(tc.flag, true)
			if got != tc.want {
				t.Errorf("resolveFormat(%q, tty=true) = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestResolveFormat_Pipe(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flag string
		want string
	}{
		{"", "jsonl"},
		{"auto", "jsonl"},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			t.Parallel()
			got := resolveFormat(tc.flag, false)
			if got != tc.want {
				t.Errorf("resolveFormat(%q, tty=false) = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestResolveFormat_FlagOverride(t *testing.T) {
	t.Parallel()

	cases := []struct {
		flag  string
		isTTY bool
		want  string
	}{
		{"table", true, "table"},
		{"raw", false, "raw"},
		{"jsonl", true, "jsonl"},
		{"json", false, "json"},
	}
	for _, tc := range cases {
		t.Run(tc.flag, func(t *testing.T) {
			t.Parallel()
			got := resolveFormat(tc.flag, tc.isTTY)
			if got != tc.want {
				t.Errorf("resolveFormat(%q, tty=%v) = %q, want %q", tc.flag, tc.isTTY, got, tc.want)
			}
		})
	}
}
