package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/config"
	"notion-cli/internal/output"
)

// NewOAuthCmd returns the parent "oauth" cobra command with subcommands.
// OAuth commands use Basic auth (client_id:client_secret), not bearer token.
func NewOAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oauth",
		Short: "OAuth token management",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		// override root's PersistentPreRunE: oauth commands don't require bearer token
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			formatFlag, _ := root.PersistentFlags().GetString("format")
			quiet, _ := root.PersistentFlags().GetBool("quiet")
			verbose, _ := root.PersistentFlags().GetBool("verbose")

			loaded, err := config.LoadNoToken(formatFlag, quiet, verbose)
			if err != nil {
				return NewCLIError(ExitAPI, err.Error())
			}
			cfg = loaded
			return nil
		},
	}
	cmd.AddCommand(newOAuthTokenCmd())
	cmd.AddCommand(newOAuthIntrospectCmd())
	cmd.AddCommand(newOAuthRevokeCmd())
	return cmd
}

func newOAuthTokenCmd() *cobra.Command {
	var code, clientID, clientSecret, redirectURI string

	cmd := &cobra.Command{
		Use:   "token",
		Short: "Exchange authorization code for an access token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOAuthToken(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format,
				clientID, clientSecret, code, redirectURI)
		},
	}
	cmd.Flags().StringVar(&code, "code", "", "Authorization code from OAuth redirect")
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret") //nolint:gosec
	cmd.Flags().StringVar(&redirectURI, "redirect-uri", "", "Redirect URI used in authorization request")
	_ = cmd.MarkFlagRequired("code")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-secret")
	return cmd
}

func newOAuthIntrospectCmd() *cobra.Command {
	var token, clientID, clientSecret string

	cmd := &cobra.Command{
		Use:   "introspect",
		Short: "Introspect an access token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOAuthIntrospect(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format,
				clientID, clientSecret, token)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Access token to introspect") //nolint:gosec
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret") //nolint:gosec
	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-secret")
	return cmd
}

func newOAuthRevokeCmd() *cobra.Command {
	var token, clientID, clientSecret string

	cmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke an access token",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runOAuthRevoke(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format,
				clientID, clientSecret, token)
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "Access token to revoke") //nolint:gosec
	cmd.Flags().StringVar(&clientID, "client-id", "", "OAuth client ID")
	cmd.Flags().StringVar(&clientSecret, "client-secret", "", "OAuth client secret") //nolint:gosec
	_ = cmd.MarkFlagRequired("token")
	_ = cmd.MarkFlagRequired("client-id")
	_ = cmd.MarkFlagRequired("client-secret")
	return cmd
}

func runOAuthToken(ctx context.Context, client *api.Client, w io.Writer, format, clientID, clientSecret, code, redirectURI string) error {
	tok, err := client.TokenExchange(ctx, clientID, clientSecret, code, redirectURI)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, tok)
}

func runOAuthIntrospect(ctx context.Context, client *api.Client, w io.Writer, format, clientID, clientSecret, token string) error {
	info, err := client.IntrospectToken(ctx, clientID, clientSecret, token)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, info)
}

func runOAuthRevoke(ctx context.Context, client *api.Client, w io.Writer, format, clientID, clientSecret, token string) error {
	result, err := client.RevokeToken(ctx, clientID, clientSecret, token)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, result)
}
