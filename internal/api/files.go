package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
)

// FileUploadCreatedBy holds the creator info for a file upload.
type FileUploadCreatedBy struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

// FileUploadParts holds the number of parts info for a multi-part upload.
type FileUploadParts struct {
	Total int `json:"total"`
	Sent  int `json:"sent"`
}

// FileUpload represents a Notion file upload object.
type FileUpload struct {
	Object           string              `json:"object"`
	ID               string              `json:"id"`
	CreatedTime      string              `json:"created_time"`
	CreatedBy        FileUploadCreatedBy `json:"created_by"`
	LastEditedTime   string              `json:"last_edited_time"`
	InTrash          bool                `json:"in_trash"`
	ExpiryTime       *string             `json:"expiry_time"`
	Status           string              `json:"status"`
	Filename         *string             `json:"filename"`
	ContentType      *string             `json:"content_type"`
	ContentLength    *int64              `json:"content_length"`
	UploadURL        string              `json:"upload_url"`
	CompleteURL      string              `json:"complete_url"`
	NumberOfParts    FileUploadParts     `json:"number_of_parts"`
	FileImportResult json.RawMessage     `json:"file_import_result,omitempty"`
}

// CreateFileUploadParams holds optional parameters for initiating a file upload.
type CreateFileUploadParams struct {
	Mode          string `json:"mode,omitempty"`
	Filename      string `json:"filename,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	NumberOfParts int    `json:"number_of_parts,omitempty"`
	ExternalURL   string `json:"external_url,omitempty"`
}

// GetFileUpload retrieves a file upload by ID (GET /v1/file_uploads/{id}).
func (c *Client) GetFileUpload(ctx context.Context, fileUploadID string) (*FileUpload, error) {
	raw, err := c.Get(ctx, "/v1/file_uploads/"+url.PathEscape(fileUploadID), nil)
	if err != nil {
		return nil, err
	}
	var fu FileUpload
	if err := json.Unmarshal(raw, &fu); err != nil {
		return nil, fmt.Errorf("decode file upload: %w", err)
	}
	return &fu, nil
}

// writeMultipartFile writes file content and optional part_number into mw, then closes pw.
func writeMultipartFile(mw *multipart.Writer, pw *io.PipeWriter, filename, contentType string, content io.Reader, partNumber int) {
	err := func() error {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filename}))
		h.Set("Content-Type", contentType)
		part, err := mw.CreatePart(h)
		if err != nil {
			return fmt.Errorf("create multipart part: %w", err)
		}
		if _, err := io.Copy(part, content); err != nil {
			return fmt.Errorf("write file content: %w", err)
		}
		if partNumber > 0 {
			if err := mw.WriteField("part_number", strconv.Itoa(partNumber)); err != nil {
				return fmt.Errorf("write part_number: %w", err)
			}
		}
		if err := mw.Close(); err != nil {
			return fmt.Errorf("close multipart writer: %w", err)
		}
		return nil
	}()
	if err != nil {
		_ = pw.CloseWithError(err)
	} else {
		_ = pw.Close()
	}
}

// SendFileContent uploads file content via multipart/form-data (POST /v1/file_uploads/{id}/send).
// partNumber is optional; pass 0 to omit it (used for single-part uploads).
// Content is streamed directly without buffering the entire file in memory.
func (c *Client) SendFileContent(ctx context.Context, fileUploadID, filename, contentType string, content io.Reader, partNumber int) (*FileUpload, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	go writeMultipartFile(mw, pw, filename, contentType, content, partNumber)

	path := "/v1/file_uploads/" + url.PathEscape(fileUploadID) + "/send"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, pr)
	if err != nil {
		_ = pr.CloseWithError(err)
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}
	var fu FileUpload
	if err := json.Unmarshal(raw, &fu); err != nil {
		return nil, fmt.Errorf("decode file upload: %w", err)
	}
	return &fu, nil
}

// CompleteFileUpload marks a file upload as complete (POST /v1/file_uploads/{id}/complete).
func (c *Client) CompleteFileUpload(ctx context.Context, fileUploadID string) (*FileUpload, error) {
	path := "/v1/file_uploads/" + url.PathEscape(fileUploadID) + "/complete"
	raw, err := c.Post(ctx, path, nil)
	if err != nil {
		return nil, err
	}
	var fu FileUpload
	if err := json.Unmarshal(raw, &fu); err != nil {
		return nil, fmt.Errorf("decode file upload: %w", err)
	}
	return &fu, nil
}

// CreateFileUpload initiates a new file upload (POST /v1/file_uploads).
func (c *Client) CreateFileUpload(ctx context.Context, params CreateFileUploadParams) (*FileUpload, error) {
	raw, err := c.Post(ctx, "/v1/file_uploads", params)
	if err != nil {
		return nil, err
	}
	var fu FileUpload
	if err := json.Unmarshal(raw, &fu); err != nil {
		return nil, fmt.Errorf("decode file upload: %w", err)
	}
	return &fu, nil
}
