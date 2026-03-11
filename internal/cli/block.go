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
	cmd.AddCommand(NewBlockChildrenCmd())
	cmd.AddCommand(NewBlockAppendCmd())
	cmd.AddCommand(NewBlockDeleteCmd())
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

// NewBlockChildrenCmd returns the "block children" cobra subcommand.
func NewBlockChildrenCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "children <block_id>",
		Short: "List child blocks of a Notion block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockChildren(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runBlockChildren(ctx context.Context, client *api.Client, w io.Writer, format, blockID string) error {
	blocks, err := client.ListBlockChildren(ctx, blockID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, blocks)
}

// NewBlockAppendCmd returns the "block append" cobra subcommand.
func NewBlockAppendCmd() *cobra.Command {
	var childrenFlag string

	cmd := &cobra.Command{
		Use:   "append <block_id>",
		Short: "Append child blocks to a Notion block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockAppend(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], childrenFlag)
		},
	}
	cmd.Flags().StringVar(&childrenFlag, "children", "[]", "Child blocks as JSON array")
	return cmd
}

func runBlockAppend(ctx context.Context, client *api.Client, w io.Writer, format, blockID, childrenJSON string) error {
	var children []map[string]any
	if err := json.Unmarshal([]byte(childrenJSON), &children); err != nil {
		return fmt.Errorf("--children: %w", err)
	}
	if len(children) == 0 {
		return fmt.Errorf("--children must be a non-empty JSON array")
	}

	blocks, err := client.AppendBlockChildren(ctx, blockID, children)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, blocks)
}

// NewBlockDeleteCmd returns the "block delete" cobra subcommand.
func NewBlockDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <block_id>",
		Short: "Delete a Notion block by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockDelete(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runBlockDelete(ctx context.Context, client *api.Client, w io.Writer, format, blockID string) error {
	block, err := client.DeleteBlock(ctx, blockID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, block)
}

func runBlockUpdate(ctx context.Context, client *api.Client, w io.Writer, format, blockID, dataJSON string, archive, unarchive bool) error {
	var typeContent map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &typeContent); err != nil {
		return fmt.Errorf("--data: %w", err)
	}

	if len(typeContent) == 0 && !archive && !unarchive {
		return fmt.Errorf("provide at least one of --data (non-empty), --archive, or --unarchive")
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
