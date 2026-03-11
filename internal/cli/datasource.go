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

// NewDSCmd returns the parent "datasource" cobra command with subcommands.
func NewDSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datasource",
		Short: "Manage Notion data sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewDSListCmd())
	cmd.AddCommand(NewDSGetCmd())
	cmd.AddCommand(NewDSCreateCmd())
	cmd.AddCommand(NewDSUpdateCmd())
	cmd.AddCommand(NewDSQueryCmd())
	cmd.AddCommand(NewDSTemplatesCmd())
	return cmd
}

// NewDSListCmd returns the "datasource list" cobra subcommand.
func NewDSListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible Notion data sources",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSList(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format)
		},
	}
}

func runDSList(ctx context.Context, client *api.Client, w io.Writer, format string) error {
	dss, err := client.ListDataSources(ctx)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, dss)
}

// NewDSGetCmd returns the "datasource get" cobra subcommand.
func NewDSGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <data_source_id>",
		Short: "Get a Notion data source by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runDSGet(ctx context.Context, client *api.Client, w io.Writer, format, dataSourceID string) error {
	ds, err := client.GetDataSource(ctx, dataSourceID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, ds)
}

// NewDSCreateCmd returns the "datasource create" cobra subcommand.
func NewDSCreateCmd() *cobra.Command {
	var parentFlag, titleFlag, propertiesFlag string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new Notion data source",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSCreate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, parentFlag, titleFlag, propertiesFlag)
		},
	}
	cmd.Flags().StringVar(&parentFlag, "parent", "", "Parent: page_id:id or workspace")
	cmd.Flags().StringVar(&titleFlag, "title", "", "Data source title (plain text)")
	cmd.Flags().StringVar(&propertiesFlag, "properties", "", "Properties as JSON object")
	_ = cmd.MarkFlagRequired("parent")
	return cmd
}

func runDSCreate(ctx context.Context, client *api.Client, w io.Writer, format, parentStr, title, propertiesJSON string) error {
	parent, err := parseDBParent(parentStr)
	if err != nil {
		return fmt.Errorf("--parent: %w", err)
	}

	req := &api.CreateDataSourceRequest{
		Parent: parent,
	}
	if title != "" {
		req.Title = []any{map[string]any{
			"type": "text",
			"text": map[string]any{"content": title},
		}}
	}
	if propertiesJSON != "" {
		props, err := parseJSONObject("--properties", propertiesJSON)
		if err != nil {
			return err
		}
		req.Properties = props
	}

	ds, err := client.CreateDataSource(ctx, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, ds)
}

// NewDSUpdateCmd returns the "datasource update" cobra subcommand.
func NewDSUpdateCmd() *cobra.Command {
	var titleFlag, descriptionFlag, propertiesFlag string

	cmd := &cobra.Command{
		Use:   "update <data_source_id>",
		Short: "Update a Notion data source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSUpdate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], titleFlag, descriptionFlag, propertiesFlag)
		},
	}
	cmd.Flags().StringVar(&titleFlag, "title", "", "New data source title (plain text)")
	cmd.Flags().StringVar(&descriptionFlag, "description", "", "New data source description (plain text)")
	cmd.Flags().StringVar(&propertiesFlag, "properties", "", "Properties as JSON object")
	return cmd
}

func runDSUpdate(ctx context.Context, client *api.Client, w io.Writer, format, dataSourceID, title, description, propertiesJSON string) error {
	if title == "" && description == "" && propertiesJSON == "" {
		return fmt.Errorf("at least one of --title, --description, --properties must be specified")
	}

	req := &api.UpdateDataSourceRequest{}
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

	ds, err := client.UpdateDataSource(ctx, dataSourceID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, ds)
}

// NewDSQueryCmd returns the "datasource query" cobra subcommand.
func NewDSQueryCmd() *cobra.Command {
	var filterFlag string

	cmd := &cobra.Command{
		Use:   "query <data_source_id>",
		Short: "Query a Notion data source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSQuery(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], filterFlag)
		},
	}
	cmd.Flags().StringVar(&filterFlag, "filter", "", "Filter as JSON object")
	return cmd
}

func runDSQuery(ctx context.Context, client *api.Client, w io.Writer, format, dataSourceID, filterJSON string) error {
	req := &api.QueryDataSourceRequest{}
	if filterJSON != "" {
		if _, err := parseJSONObject("--filter", filterJSON); err != nil {
			return err
		}
		req.Filter = json.RawMessage(filterJSON)
	}

	results, err := client.QueryDataSource(ctx, dataSourceID, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, results)
}

// NewDSTemplatesCmd returns the "datasource templates" cobra subcommand.
func NewDSTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates <data_source_id>",
		Short: "Get templates for a Notion data source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDSTemplates(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runDSTemplates(ctx context.Context, client *api.Client, w io.Writer, format, dataSourceID string) error {
	templates, err := client.GetDataSourceTemplates(ctx, dataSourceID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, templates)
}
