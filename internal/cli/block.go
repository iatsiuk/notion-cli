package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewBlockCmd returns the parent "block" cobra command with subcommands.
func NewBlockCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block",
		Short: "Manage Notion blocks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewBlockGetCmd())
	return cmd
}

// NewBlockGetCmd returns the "block get" cobra subcommand.
func NewBlockGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <block_id>",
		Short: "Get a Notion block by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runBlockGet(ctx context.Context, client *api.Client, w io.Writer, format, blockID string) error {
	block, err := client.GetBlock(ctx, blockID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, block)
}
