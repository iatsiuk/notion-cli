package cli

import (
	"context"
	"encoding/json"
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
	cmd.AddCommand(NewBlockUpdateCmd())
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

// NewBlockUpdateCmd returns the "block update" cobra subcommand.
func NewBlockUpdateCmd() *cobra.Command {
	var dataFlag string
	var archiveFlag, unarchiveFlag bool

	cmd := &cobra.Command{
		Use:   "update <block_id>",
		Short: "Update a Notion block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockUpdate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], dataFlag, archiveFlag, unarchiveFlag)
		},
	}
	cmd.Flags().StringVar(&dataFlag, "data", "{}", "Block type content as JSON object")
	cmd.Flags().BoolVar(&archiveFlag, "archive", false, "Archive the block")
	cmd.Flags().BoolVar(&unarchiveFlag, "unarchive", false, "Unarchive the block")
	cmd.MarkFlagsMutuallyExclusive("archive", "unarchive")
	return cmd
}

func runBlockUpdate(ctx context.Context, client *api.Client, w io.Writer, format, blockID, dataJSON string, archive, unarchive bool) error {
	var typeContent map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &typeContent); err != nil {
		return fmt.Errorf("--data: %w", err)
	}

	req := &api.UpdateBlockRequest{}
	if len(typeContent) > 0 {
		req.TypeContent = typeContent
	}
	if archive {
		t := true
		req.Archived = &t
	} else if unarchive {
		f := false
		req.Archived = &f
	}

	block, err := client.UpdateBlock(ctx, blockID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, block)
}
