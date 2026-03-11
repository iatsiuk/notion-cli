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
	cmd.AddCommand(NewDBCreateCmd())
	cmd.AddCommand(NewDBUpdateCmd())
	cmd.AddCommand(NewDBQueryCmd())
	return cmd
}

// parseJSONObject parses a JSON string and validates it is a JSON object (not null).
func parseJSONObject(flag, s string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, fmt.Errorf("%s: %w", flag, err)
	}
	if m == nil {
		return nil, fmt.Errorf("%s: must be a JSON object, got null", flag)
	}
	return m, nil
}

// parseDBParent parses parent string for database create.
// Accepts "page_id:id" or "workspace".
func parseDBParent(s string) (api.Parent, error) {
	if s == "workspace" {
		return api.Parent{Type: "workspace", Workspace: true}, nil
	}
	idx := strings.Index(s, ":")
	if idx < 0 {
		return api.Parent{}, fmt.Errorf("invalid parent format %q: expected page_id:id or workspace", s)
	}
	typ, id := s[:idx], s[idx+1:]
	if id == "" {
		return api.Parent{}, fmt.Errorf("invalid parent format %q: id is empty", s)
	}
	if typ != "page_id" {
		return api.Parent{}, fmt.Errorf("unsupported parent type %q: use page_id or workspace", typ)
	}
	return api.Parent{Type: "page_id", PageID: id}, nil
}

// NewDBCreateCmd returns the "db create" cobra subcommand.
func NewDBCreateCmd() *cobra.Command {
	var parentFlag, titleFlag, propertiesFlag string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Notion database",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBCreate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, parentFlag, titleFlag, propertiesFlag)
		},
	}
	cmd.Flags().StringVar(&parentFlag, "parent", "", "Parent: page_id:id or workspace")
	cmd.Flags().StringVar(&titleFlag, "title", "", "Database title (plain text)")
	cmd.Flags().StringVar(&propertiesFlag, "properties", "{}", "Properties schema as JSON object")
	_ = cmd.MarkFlagRequired("parent")
	return cmd
}

func runDBCreate(ctx context.Context, client *api.Client, w io.Writer, format, parentStr, title, propertiesJSON string) error {
	parent, err := parseDBParent(parentStr)
	if err != nil {
		return fmt.Errorf("--parent: %w", err)
	}

	props, err := parseJSONObject("--properties", propertiesJSON)
	if err != nil {
		return err
	}

	req := &api.CreateDatabaseRequest{
		Parent: parent,
	}
	if title != "" {
		req.Title = []any{map[string]any{
			"type": "text",
			"text": map[string]any{"content": title},
		}}
	}
	req.Properties = props

	db, err := client.CreateDatabase(ctx, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, db)
}

// NewDBUpdateCmd returns the "db update" cobra subcommand.
func NewDBUpdateCmd() *cobra.Command {
	var titleFlag, descriptionFlag, propertiesFlag string

	cmd := &cobra.Command{
		Use:   "update <database_id>",
		Short: "Update a Notion database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBUpdate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], titleFlag, descriptionFlag, propertiesFlag)
		},
	}
	cmd.Flags().StringVar(&titleFlag, "title", "", "New database title (plain text)")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "New database description (plain text)")
	cmd.Flags().StringVar(&propertiesFlag, "properties", "", "Properties schema as JSON object")
	return cmd
}

func runDBUpdate(ctx context.Context, client *api.Client, w io.Writer, format, databaseID, title, description, propertiesJSON string) error {
	if title == "" && description == "" && propertiesJSON == "" {
		return fmt.Errorf("at least one of --title, --description, --properties must be specified")
	}
	req := &api.UpdateDatabaseRequest{}

	if title != "" {
		req.Title = []any{map[string]any{
			"type": "text",
			"text": map[string]any{"content": title},
		}}
	}
	if description != "" {
		req.Description = []any{map[string]any{
			"type": "text",
			"text": map[string]any{"content": description},
		}}
	}
	if propertiesJSON != "" {
		props, err := parseJSONObject("--properties", propertiesJSON)
		if err != nil {
			return err
		}
		req.Properties = props
	}

	db, err := client.UpdateDatabase(ctx, databaseID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, db)
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

// NewDBQueryCmd returns the "db query" cobra subcommand.
func NewDBQueryCmd() *cobra.Command {
	var filterFlag, sortFlag string

	cmd := &cobra.Command{
		Use:   "query <database_id>",
		Short: "Query a Notion database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDBQuery(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], filterFlag, sortFlag)
		},
	}
	cmd.Flags().StringVar(&filterFlag, "filter", "", "Filter as JSON object")
	cmd.Flags().StringVar(&sortFlag, "sort", "", "Sorts as JSON array")
	return cmd
}

func runDBQuery(ctx context.Context, client *api.Client, w io.Writer, format, databaseID, filterJSON, sortJSON string) error {
	req := &api.QueryDatabaseRequest{}

	if filterJSON != "" {
		var filterObj map[string]any
		if err := json.Unmarshal([]byte(filterJSON), &filterObj); err != nil {
			return fmt.Errorf("--filter: must be a JSON object")
		}
		if filterObj == nil {
			return fmt.Errorf("--filter: must be a JSON object, got null")
		}
		req.Filter = json.RawMessage(filterJSON)
	}
	if sortJSON != "" {
		var sortArr []any
		if err := json.Unmarshal([]byte(sortJSON), &sortArr); err != nil {
			return fmt.Errorf("--sort: must be a JSON array")
		}
		if sortArr == nil {
			return fmt.Errorf("--sort: must be a JSON array, got null")
		}
		req.Sorts = json.RawMessage(sortJSON)
	}

	pages, err := client.QueryDatabase(ctx, databaseID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, pages)
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
