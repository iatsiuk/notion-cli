package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewDBCmd returns the parent "db" cobra command with subcommands.
func NewDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage Notion databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewDBGetCmd())
	return cmd
}

// NewDBGetCmd returns the "db get" cobra subcommand.
func NewDBGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <database_id>",
		Short: "Get a Notion database by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runDBGet(ctx context.Context, client *api.Client, w io.Writer, format, databaseID string) error {
	db, err := client.GetDatabase(ctx, databaseID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, db)
}
