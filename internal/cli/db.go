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
	cmd.AddCommand(NewDBListCmd())
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

// NewDBListCmd returns the "db list" cobra subcommand.
func NewDBListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible Notion databases",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBList(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format)
		},
	}
}

func runDBList(ctx context.Context, client *api.Client, w io.Writer, format string) error {
	dbs, err := client.ListDatabases(ctx)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, dbs)
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
