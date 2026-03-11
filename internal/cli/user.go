package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// isCharDevice reports whether v (*os.File) is a TTY.
func isCharDevice(v any) bool {
	f, ok := v.(*os.File)
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd())) //nolint:gosec
}

// isTerminal reports whether w is a character device (TTY).
func isTerminal(w io.Writer) bool {
	return isCharDevice(w)
}

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
	cmd.AddCommand(NewUserListCmd())
	cmd.AddCommand(NewUserGetCmd())
	return cmd
}

// newClientFromCfg builds an API client from the global config.
func newClientFromCfg() *api.Client {
	opts := []api.Option{api.WithHTTPClient(&http.Client{Timeout: 30 * time.Second})}
	if cfg.Verbose {
		opts = append(opts, api.WithVerbose(os.Stderr))
	}
	return api.NewClient(cfg.Token, opts...)
}

// NewUserMeCmd returns the "user me" cobra subcommand.
func NewUserMeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "me",
		Short: "Get the currently authenticated user",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserMe(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format)
		},
	}
}

// NewUserListCmd returns the "user list" cobra subcommand.
func NewUserListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all workspace users",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserList(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format)
		},
	}
}

func runUserList(ctx context.Context, client *api.Client, w io.Writer, format string) error {
	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	all := make([]api.User, 0)
	err = client.ListUsers(ctx, func(users []api.User) error {
		all = append(all, users...)
		return nil
	})
	if err != nil {
		return mapAPIError(err)
	}
	return f.Format(w, all)
}

// NewUserGetCmd returns the "user get" cobra subcommand.
func NewUserGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <user_id>",
		Short: "Get a workspace user by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUserGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runUserGet(ctx context.Context, client *api.Client, w io.Writer, format, userID string) error {
	user, err := client.GetUser(ctx, userID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, user)
}

func runUserMe(ctx context.Context, client *api.Client, w io.Writer, format string) error {
	user, err := client.GetMe(ctx)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, user)
}

// mapAPIError converts api errors to CLIError with appropriate exit codes.
func mapAPIError(err error) error {
	var apiErr *api.APIError
	if api.AsAPIError(err, &apiErr) {
		return NewCLIError(apiErr.ExitCode(), apiErr.Error())
	}
	var connErr *api.ConnectionError
	if api.AsConnectionError(err, &connErr) {
		return NewCLIError(ExitConnection, connErr.Error())
	}
	return err
}
