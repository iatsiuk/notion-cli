package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	"notion-cli/internal/config"
)

const notionAPIBase = "https://api.notion.com"
const notionVersion = "2022-06-28"

// NewStatusCmd returns a cobra command that checks connectivity to the Notion API.
func NewStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check connectivity to the Notion API",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(cmd.Context(), cfg, notionAPIBase, http.DefaultClient)
		},
	}
}

// runStatus performs GET /v1/users/me and maps the response to a CLIError.
func runStatus(ctx context.Context, c *config.Config, baseURL string, client *http.Client) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/v1/users/me", http.NoBody)
	if err != nil {
		return NewCLIError(ExitConnection, fmt.Sprintf("build request: %v", err))
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Notion-Version", notionVersion)

	resp, err := client.Do(req)
	if err != nil {
		return NewCLIError(ExitConnection, fmt.Sprintf("connection error: %v", err))
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		if !c.Quiet {
			fmt.Println("ok")
		}
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		return NewCLIError(ExitAuth, fmt.Sprintf("auth error: HTTP %d", resp.StatusCode))
	default:
		return NewCLIError(ExitAPI, fmt.Sprintf("API error: HTTP %d", resp.StatusCode))
	}
}
