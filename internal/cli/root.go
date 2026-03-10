package cli

import (
	"github.com/spf13/cobra"

	"notion-cli/internal/config"
)

// cfg holds the loaded config after PersistentPreRunE runs.
var cfg *config.Config

// Config returns the loaded configuration. Valid only after PersistentPreRunE.
func Config() *config.Config { return cfg }

type rootFlags struct {
	token   string
	format  string
	quiet   bool
	verbose bool
}

// NewRootCmd returns the root cobra command with global flags registered.
func NewRootCmd() *cobra.Command {
	f := &rootFlags{}

	cmd := &cobra.Command{
		Use:           "notion-cli",
		Short:         "Notion CLI - interact with the Notion API",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	p := cmd.PersistentFlags()
	p.StringVar(&f.token, "token", "", "Notion API token (overrides NOTION_TOKEN env var)")
	p.StringVar(&f.format, "format", "auto", "Output format: json|jsonl|raw|table|auto")
	p.BoolVar(&f.quiet, "quiet", false, "Suppress non-essential output")
	p.BoolVar(&f.verbose, "verbose", false, "Enable verbose output")

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		loaded, err := config.Load(f.token, f.format, f.quiet, f.verbose)
		if err != nil {
			return err
		}
		cfg = loaded
		return nil
	}

	cmd.AddCommand(NewStatusCmd())

	return cmd
}

// ExecuteRootForTest executes cmd and returns the loaded config. For use in tests only.
func ExecuteRootForTest(cmd *cobra.Command) (*config.Config, error) {
	cfg = nil
	err := cmd.Execute()
	return cfg, err
}
