package cli_test

import (
	"errors"
	"testing"

	"notion-cli/internal/cli"
)

func TestNewRootCmd_HasGlobalFlags(t *testing.T) {
	t.Parallel()
	cmd := cli.NewRootCmd()
	flags := cmd.PersistentFlags()

	for _, name := range []string{"token", "format", "quiet", "verbose"} {
		if flags.Lookup(name) == nil {
			t.Errorf("expected flag --%s to be registered", name)
		}
	}
}

func TestNewRootCmd_TokenFlagOverridesEnv(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "env-token")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--token", "flag-token"})

	cfg, err := cli.ExecuteRootForTest(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "flag-token" {
		t.Errorf("expected token=flag-token, got %s", cfg.Token)
	}
}

func TestNewRootCmd_TokenFromEnv(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "env-token")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{})

	cfg, err := cli.ExecuteRootForTest(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "env-token" {
		t.Errorf("expected token=env-token, got %s", cfg.Token)
	}
}

func TestNewRootCmd_NoTokenReturnsError(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{})

	_, err := cli.ExecuteRootForTest(cmd)
	if err == nil {
		t.Fatal("expected error when no token provided")
	}
	var cliErr *cli.CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != cli.ExitAuth {
		t.Errorf("expected exit code %d, got %d", cli.ExitAuth, cliErr.Code)
	}
}

func TestNewRootCmd_FormatFlag(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "tok")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--format", "jsonl"})

	cfg, err := cli.ExecuteRootForTest(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Format != "jsonl" {
		t.Errorf("expected format=jsonl, got %s", cfg.Format)
	}
}

func TestNewRootCmd_QuietFlag(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "tok")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--quiet"})

	cfg, err := cli.ExecuteRootForTest(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Quiet {
		t.Error("expected quiet=true")
	}
}

func TestNewRootCmd_VerboseFlag(t *testing.T) {
	t.Setenv("NOTION_TOKEN", "tok")

	cmd := cli.NewRootCmd()
	cmd.SetArgs([]string{"--verbose"})

	cfg, err := cli.ExecuteRootForTest(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Verbose {
		t.Error("expected verbose=true")
	}
}
