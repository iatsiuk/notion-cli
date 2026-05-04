package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

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
	var childrenFlag, afterFlag string

	cmd := &cobra.Command{
		Use:   "append <block_id>",
		Short: "Append child blocks to a Notion block",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockAppend(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cmd.InOrStdin(), cfg.Format, args[0], childrenFlag, cmd.Flags().Changed("children"), afterFlag, cmd.Flags().Changed("after"))
		},
	}
	cmd.Flags().StringVar(&childrenFlag, "children", "[]", "Child blocks as JSON array, or full request body as JSON object {\"children\":[...], \"after\":\"...\"}")
	cmd.Flags().StringVar(&afterFlag, "after", "", "Insert new children after this block ID (default: append to end)")
	return cmd
}

func runBlockAppend(ctx context.Context, client *api.Client, w io.Writer, stdin io.Reader, format, blockID, childrenJSON string, childrenFlagSet bool, afterFlag string, afterFlagSet bool) error {
	if !childrenFlagSet {
		fromStdin, err := readChildrenFromStdin(stdin)
		if err != nil {
			return err
		}
		childrenJSON = fromStdin
	}

	children, afterFromJSON, hadAfterKey, err := parseAppendChildren(childrenJSON)
	if err != nil {
		return err
	}
	if afterFlagSet && hadAfterKey {
		return fmt.Errorf("conflicting after: provided via --after and inside --children JSON object")
	}
	if len(children) == 0 {
		return fmt.Errorf("children must be a non-empty JSON array")
	}

	after := afterFromJSON
	if afterFlagSet {
		after = afterFlag
	}

	blocks, err := client.AppendBlockChildren(ctx, blockID, children, after)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, blocks)
}

func readChildrenFromStdin(stdin io.Reader) (string, error) {
	if isTTY(stdin) {
		return "", fmt.Errorf("stdin is a terminal: provide --children flag or pipe JSON array via stdin")
	}
	data, err := io.ReadAll(stdin)
	if err != nil {
		return "", fmt.Errorf("reading stdin: %w", err)
	}
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return "", fmt.Errorf("provide --children flag or pipe JSON array via stdin")
	}
	return trimmed, nil
}

func validateChildrenElements(children []map[string]any) error {
	for i, c := range children {
		if c == nil {
			return fmt.Errorf("--children: element %d is not a JSON object", i)
		}
	}
	return nil
}

// parseAppendChildren accepts either a JSON array of child block objects or a
// full request body object of the form {"children":[...], "after":"..."}.
// hadAfterKey reflects whether the "after" key was present in the object form,
// independently of whether its value is the empty string.
func parseAppendChildren(raw string) (children []map[string]any, afterFromJSON string, hadAfterKey bool, err error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, "", false, fmt.Errorf("--children: empty input")
	}

	var asObject map[string]any
	if objErr := json.Unmarshal([]byte(trimmed), &asObject); objErr == nil {
		return parseAppendChildrenObject(asObject)
	}

	if arrErr := json.Unmarshal([]byte(trimmed), &children); arrErr != nil {
		return nil, "", false, fmt.Errorf("--children: %w", arrErr)
	}
	if err := validateChildrenElements(children); err != nil {
		return nil, "", false, err
	}
	return children, "", false, nil
}

func parseAppendChildrenObject(obj map[string]any) (children []map[string]any, afterFromJSON string, hadAfterKey bool, err error) {
	for k := range obj {
		if k != "children" && k != "after" {
			return nil, "", false, fmt.Errorf("--children: unknown key %q in object form (allowed: children, after)", k)
		}
	}

	rawChildren, ok := obj["children"]
	if !ok {
		return nil, "", false, fmt.Errorf("--children: object form requires \"children\" key")
	}

	encoded, err := json.Marshal(rawChildren)
	if err != nil {
		return nil, "", false, fmt.Errorf("--children: %w", err)
	}
	if err := json.Unmarshal(encoded, &children); err != nil {
		return nil, "", false, fmt.Errorf("--children: %w", err)
	}
	if err := validateChildrenElements(children); err != nil {
		return nil, "", false, err
	}

	afterValue, hadAfterKey := obj["after"]
	if hadAfterKey {
		s, ok := afterValue.(string)
		if !ok {
			return nil, "", false, fmt.Errorf("--children: \"after\" must be a string")
		}
		afterFromJSON = s
	}
	return children, afterFromJSON, hadAfterKey, nil
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
		req.InTrash = &t
	} else if unarchive {
		f := false
		req.InTrash = &f
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
