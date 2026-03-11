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
	cmd.AddCommand(NewPageCreateCmd())
	cmd.AddCommand(NewPageUpdateCmd())
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

// NewPageCreateCmd returns the "page create" cobra subcommand.
func NewPageCreateCmd() *cobra.Command {
	var parentFlag, propertiesFlag string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Notion page",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPageCreate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, parentFlag, propertiesFlag)
		},
	}
	cmd.Flags().StringVar(&parentFlag, "parent", "", "Parent: type:id (e.g. database_id:abc or page_id:abc)")
	cmd.Flags().StringVar(&propertiesFlag, "properties", "{}", "Properties as JSON object")
	_ = cmd.MarkFlagRequired("parent")
	return cmd
}

// parseParent parses "type:id" into a Parent struct.
// Supported types: database_id, page_id.
func parseParent(s string) (api.Parent, error) {
	idx := strings.Index(s, ":")
	if idx < 0 {
		return api.Parent{}, fmt.Errorf("invalid parent format %q: expected type:id", s)
	}
	typ, id := s[:idx], s[idx+1:]
	switch typ {
	case "database_id":
		return api.Parent{Type: "database_id", DatabaseID: id}, nil
	case "page_id":
		return api.Parent{Type: "page_id", PageID: id}, nil
	default:
		return api.Parent{}, fmt.Errorf("unsupported parent type %q: use database_id or page_id", typ)
	}
}

// NewPageUpdateCmd returns the "page update" cobra subcommand.
func NewPageUpdateCmd() *cobra.Command {
	var propertiesFlag string
	var archiveFlag, unarchiveFlag bool

	cmd := &cobra.Command{
		Use:   "update <page_id>",
		Short: "Update a Notion page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPageUpdate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], propertiesFlag, archiveFlag, unarchiveFlag)
		},
	}
	cmd.Flags().StringVar(&propertiesFlag, "properties", "{}", "Properties as JSON object")
	cmd.Flags().BoolVar(&archiveFlag, "archive", false, "Archive the page")
	cmd.Flags().BoolVar(&unarchiveFlag, "unarchive", false, "Unarchive the page")
	return cmd
}

func runPageUpdate(ctx context.Context, client *api.Client, w io.Writer, format, pageID, propertiesJSON string, archive, unarchive bool) error {
	var props map[string]any
	if err := json.Unmarshal([]byte(propertiesJSON), &props); err != nil {
		return fmt.Errorf("--properties: %w", err)
	}

	req := &api.UpdatePageRequest{Properties: props}
	if archive {
		t := true
		req.Archived = &t
	} else if unarchive {
		f := false
		req.Archived = &f
	}

	page, err := client.UpdatePage(ctx, pageID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, page)
}

func runPageCreate(ctx context.Context, client *api.Client, w io.Writer, format, parentStr, propertiesJSON string) error {
	parent, err := parseParent(parentStr)
	if err != nil {
		return fmt.Errorf("--parent: %w", err)
	}

	var props map[string]any
	if err := json.Unmarshal([]byte(propertiesJSON), &props); err != nil {
		return fmt.Errorf("--properties: %w", err)
	}

	req := &api.CreatePageRequest{
		Parent:     parent,
		Properties: props,
	}
	page, err := client.CreatePage(ctx, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, page)
}
