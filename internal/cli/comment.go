package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewCommentCmd returns the parent "comment" cobra command with subcommands.
func NewCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage Notion comments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewCommentListCmd())
	return cmd
}

// NewCommentListCmd returns the "comment list" cobra subcommand.
func NewCommentListCmd() *cobra.Command {
	var blockID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments on a block or page",
		RunE: func(cmd *cobra.Command, args []string) error {
			if blockID == "" {
				return fmt.Errorf("--block flag is required")
			}
			return runCommentList(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, blockID)
		},
	}
	cmd.Flags().StringVar(&blockID, "block", "", "Block or page ID to list comments for")
	return cmd
}

func runCommentList(ctx context.Context, client *api.Client, w io.Writer, format, blockID string) error {
	comments, err := client.ListComments(ctx, blockID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, comments)
}
