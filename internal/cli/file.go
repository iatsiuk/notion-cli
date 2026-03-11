package cli

import (
	"context"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"notion-cli/internal/api"
	"notion-cli/internal/output"
)

// NewFileCmd returns the parent "file" cobra command with subcommands.
func NewFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Manage Notion file uploads",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(NewFileCreateCmd())
	cmd.AddCommand(NewFileGetCmd())
	cmd.AddCommand(NewFileSendCmd())
	cmd.AddCommand(NewFileCompleteCmd())
	cmd.AddCommand(NewFileUploadCmd())
	return cmd
}

// NewFileCreateCmd returns the "file create" cobra subcommand.
func NewFileCreateCmd() *cobra.Command {
	var filenameFlag, contentTypeFlag, modeFlag string
	var numPartsFlag int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Initiate a new file upload",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileCreate(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, filenameFlag, contentTypeFlag, modeFlag, numPartsFlag)
		},
	}
	cmd.Flags().StringVar(&filenameFlag, "filename", "", "Filename for the upload")
	cmd.Flags().StringVar(&contentTypeFlag, "content-type", "", "MIME content type")
	cmd.Flags().StringVar(&modeFlag, "mode", "", "Upload mode (e.g. multi)")
	cmd.Flags().IntVar(&numPartsFlag, "number-of-parts", 0, "Number of parts for multipart upload")
	return cmd
}

func runFileCreate(ctx context.Context, client *api.Client, w io.Writer, format, filename, contentType, mode string, numParts int) error {
	params := api.CreateFileUploadParams{
		Filename:      filename,
		ContentType:   contentType,
		Mode:          mode,
		NumberOfParts: numParts,
	}
	fu, err := client.CreateFileUpload(ctx, params)
	if err != nil {
		return mapAPIError(err)
	}
	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, fu)
}

// NewFileGetCmd returns the "file get" cobra subcommand.
func NewFileGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <file_upload_id>",
		Short: "Get a file upload by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileGet(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runFileGet(ctx context.Context, client *api.Client, w io.Writer, format, fileUploadID string) error {
	fu, err := client.GetFileUpload(ctx, fileUploadID)
	if err != nil {
		return mapAPIError(err)
	}
	f, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return f.Format(w, fu)
}

// NewFileSendCmd returns the "file send" cobra subcommand.
func NewFileSendCmd() *cobra.Command {
	var contentTypeFlag string
	var partNumberFlag int

	cmd := &cobra.Command{
		Use:   "send <file_upload_id> <file_path>",
		Short: "Upload file content to a file upload",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileSend(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0], args[1], contentTypeFlag, partNumberFlag)
		},
	}
	cmd.Flags().StringVar(&contentTypeFlag, "content-type", "", "MIME content type (auto-detected if omitted)")
	cmd.Flags().IntVar(&partNumberFlag, "part", 0, "Part number for multipart upload (0 = single part)")
	return cmd
}

func runFileSend(ctx context.Context, client *api.Client, w io.Writer, format, fileUploadID, filePath, contentType string, partNumber int) error {
	if partNumber < 0 {
		return fmt.Errorf("part number must be >= 0")
	}
	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	filename := filepath.Base(filePath)
	if contentType == "" {
		contentType = mime.TypeByExtension(filepath.Ext(filePath))
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	fu, err := client.SendFileContent(ctx, fileUploadID, filename, contentType, f, partNumber)
	if err != nil {
		return mapAPIError(err)
	}
	out, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return out.Format(w, fu)
}

// NewFileCompleteCmd returns the "file complete" cobra subcommand.
func NewFileCompleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <file_upload_id>",
		Short: "Mark a file upload as complete",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileComplete(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runFileComplete(ctx context.Context, client *api.Client, w io.Writer, format, fileUploadID string) error {
	fu, err := client.CompleteFileUpload(ctx, fileUploadID)
	if err != nil {
		return mapAPIError(err)
	}
	out, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return out.Format(w, fu)
}

// NewFileUploadCmd returns the "file upload" cobra subcommand (create + send + complete).
func NewFileUploadCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upload <file_path>",
		Short: "Upload a file in one step (create, send, complete)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFileUpload(cmd.Context(), newClientFromCfg(), cmd.OutOrStdout(), cfg.Format, args[0])
		},
	}
}

func runFileUpload(ctx context.Context, client *api.Client, w io.Writer, format, filePath string) error {
	filename := filepath.Base(filePath)
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	f, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	fu, err := client.CreateFileUpload(ctx, api.CreateFileUploadParams{
		Filename:    filename,
		ContentType: contentType,
	})
	if err != nil {
		return mapAPIError(err)
	}

	uploadID := fu.ID
	sent, err := client.SendFileContent(ctx, uploadID, filename, contentType, f, 0)
	if err != nil {
		return fmt.Errorf("send file (upload ID: %s): %w", uploadID, mapAPIError(err))
	}

	if sent.Status != "uploaded" {
		fu, err = client.CompleteFileUpload(ctx, uploadID)
		if err != nil {
			return fmt.Errorf("complete upload (upload ID: %s): %w", uploadID, mapAPIError(err))
		}
	} else {
		fu = sent
	}

	out, err := output.New(format, isTerminal(w))
	if err != nil {
		return fmt.Errorf("output format: %w", err)
	}
	return out.Format(w, fu)
}
