package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewUserCmd returns the parent "user" cobra command with subcommands.
func NewUserCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage Notion workspace users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewUserMeCmd())
	return cmd
}

// NewUserMeCmd returns the "user me" cobra subcommand.
func NewUserMeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Get the currently authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient(cfg.Token,
				api.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
			)
			if cfg.Verbose {
				client = api.NewClient(cfg.Token,
					api.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
					api.WithVerbose(os.Stderr),
				)
			}
			return runUserMe(cmd.Context(), client, cmd.OutOrStdout(), cfg.Format)
		},
	}
}

func runUserMe(ctx context.Context, client *api.Client, w io.Writer, format string) error {
	user, err := client.GetMe(ctx)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, false)
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, user)
}

// mapAPIError converts api errors to CLIError with appropriate exit codes.
func mapAPIError(err error) error {
	var apiErr *api.APIError
	if api.AsAPIError(err, &apiErr) {
		code := ExitAPI
		if apiErr.Status == 401 || apiErr.Status == 403 {
			code = ExitAuth
		}
		return NewCLIError(code, apiErr.Error())
	}
	var connErr *api.ConnectionError
	if api.AsConnectionError(err, &connErr) {
		return NewCLIError(ExitConnection, connErr.Error())
	}
	return err
}
