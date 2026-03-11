package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewCommentCmd returns the parent "comment" cobra command with subcommands.
func NewCommentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage Notion comments",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewCommentListCmd())
	cmd.AddCommand(NewCommentCreateCmd())
	cmd.AddCommand(NewCommentGetCmd())
	return cmd
}

// NewCommentListCmd returns the "comment list" cobra subcommand.
func NewCommentListCmd() *cobra.Command {
	var blockID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List comments on a block or page",
		RunE: func(cmd *cobra.Command, args []string) error {
			if blockID == "" {
				return fmt.Errorf("--block flag is required")
			}
			return runCommentList(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, blockID)
		},
	}
	cmd.Flags().StringVar(&blockID, "block", "", "Block or page ID to list comments for")
	return cmd
}

// NewCommentCreateCmd returns the "comment create" cobra subcommand.
func NewCommentCreateCmd() *cobra.Command {
	var pageID, discussionID, text string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a comment on a page or in a discussion",
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageID == "" && discussionID == "" {
				return fmt.Errorf("one of --page or --discussion is required")
			}
			if text == "" {
				return fmt.Errorf("--text flag is required")
			}
			return runCommentCreate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, pageID, discussionID, text)
		},
	}
	cmd.Flags().StringVar(&pageID, "page", "", "Page ID to comment on")
	cmd.Flags().StringVar(&discussionID, "discussion", "", "Discussion ID to reply in")
	cmd.Flags().StringVar(&text, "text", "", "Comment text content")
	cmd.MarkFlagsMutuallyExclusive("page", "discussion")
	return cmd
}

func runCommentCreate(ctx context.Context, client *api.Client, w io.Writer, format, pageID, discussionID, text string) error {
	req := &api.CreateCommentRequest{
		RichText: []api.RichTextItem{{Type: "text", Text: &api.RichTextText{Content: text}}},
	}
	if pageID != "" {
		req.Parent = &api.Parent{Type: "page_id", PageID: pageID}
	} else {
		req.DiscussionID = discussionID
	}

	cmt, err := client.CreateComment(ctx, req)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, cmt)
}

// NewCommentGetCmd returns the "comment get" cobra subcommand.
func NewCommentGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <comment_id>",
		Short: "Get a Notion comment by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommentGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runCommentGet(ctx context.Context, client *api.Client, w io.Writer, format, commentID string) error {
	cmt, err := client.GetComment(ctx, commentID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, cmt)
}

func runCommentList(ctx context.Context, client *api.Client, w io.Writer, format, blockID string) error {
	comments, err := client.ListComments(ctx, blockID)
	if err != nil {
		return mapAPIError(err)
	}

	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, comments)
}
