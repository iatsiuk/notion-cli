package cli

import (
	"errors"

	"github.com/spf13/cobra"

	"notion-cli/internal/config"
)

// cfg holds the loaded config after PersistentPreRunE runs.
var cfg *config.Config

// NewRootCmd returns the root cobra command with global flags registered.
func NewRootCmd() *cobra.Command {
	var token, format string
	var quiet, verbose bool

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
	p.StringVarP(&token, "token", "t", "", "Notion API token (overrides NOTION_TOKEN env var)")
	p.StringVarP(&format, "format", "f", "auto", "Output format: json|jsonl|raw|table|auto")
	p.BoolVar(&quiet, "quiet", false, "Suppress non-essential output")
	p.BoolVar(&verbose, "verbose", false, "Enable verbose output")

	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		loaded, err := config.Load(token, format, quiet, verbose)
		if err != nil {
			if errors.Is(err, config.ErrNoToken) {
				// root command without a subcommand shows help; don't require auth
				if !cmd.HasParent() {
					return nil
				}
				return NewCLIError(ExitAuth, err.Error())
			}
			return NewCLIError(ExitAPI, err.Error())
		}
		cfg = loaded
		return nil
	}

	cmd.AddCommand(NewStatusCmd())

	return cmd
}
