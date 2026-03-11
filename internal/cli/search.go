package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewSearchCmd returns the "search" cobra command.
func NewSearchCmd() *cobra.Command {
	var typeFlag, sortFlag string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Notion pages and databases",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] == "" {
				return fmt.Errorf("query must not be empty")
			}
			return runSearch(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], typeFlag, sortFlag)
		},
	}
	cmd.Flags().StringVar(&typeFlag, "type", "", "Filter by object type: page or database")
	cmd.Flags().StringVar(&sortFlag, "sort", "", "Sort direction: ascending or descending")
	return cmd
}

func runSearch(ctx context.Context, client *api.Client, w io.Writer, format, query, typeFilter, sortDir string) error {
	req := &api.SearchRequest{Query: query}

	if typeFilter != "" {
		if typeFilter != "page" && typeFilter != "database" {
			return fmt.Errorf("--type must be page or database")
		}
		req.Filter = &api.SearchFilter{Value: typeFilter, Property: "object"}
	}

	if sortDir != "" {
		if sortDir != "ascending" && sortDir != "descending" {
			return fmt.Errorf("--sort must be ascending or descending")
		}
		req.Sort = &api.SearchSort{Direction: sortDir, Timestamp: "last_edited_time"}
	}

	results, err := client.Search(ctx, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, results)
}
