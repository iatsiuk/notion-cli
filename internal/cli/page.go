package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewPageCmd returns the parent "page" cobra command with subcommands.
func NewPageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "page",
		Short: "Manage Notion pages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewPageGetCmd())
	return cmd
}

// NewPageGetCmd returns the "page get" cobra subcommand.
func NewPageGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <page_id>",
		Short: "Get a Notion page by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPageGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runPageGet(ctx context.Context, client *api.Client, w io.Writer, format, pageID string) error {
	page, err := client.GetPage(ctx, pageID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, page)
}
