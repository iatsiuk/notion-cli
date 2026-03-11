package api

import (
	"context"
	"encoding/json"
	"fmt"
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
	Object         string              `json:"object"`
	ID             string              `json:"id"`
	CreatedTime    string              `json:"created_time"`
	CreatedBy      FileUploadCreatedBy `json:"created_by"`
	LastEditedTime string              `json:"last_edited_time"`
	InTrash        bool                `json:"in_trash"`
	ExpiryTime     *string             `json:"expiry_time"`
	Status         string              `json:"status"`
	Filename       *string             `json:"filename"`
	ContentType    *string             `json:"content_type"`
	ContentLength  *int64              `json:"content_length"`
	UploadURL      string              `json:"upload_url"`
	CompleteURL    string              `json:"complete_url"`
	NumberOfParts  FileUploadParts     `json:"number_of_parts"`
}

// CreateFileUploadParams holds optional parameters for initiating a file upload.
type CreateFileUploadParams struct {
	Mode          string `json:"mode,omitempty"`
	Filename      string `json:"filename,omitempty"`
	ContentType   string `json:"content_type,omitempty"`
	NumberOfParts int    `json:"number_of_parts,omitempty"`
	ExternalURL   string `json:"external_url,omitempty"`
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
