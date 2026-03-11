package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// PartialUser is a minimal user reference (id + object).
type PartialUser struct {
	Object string `json:"object"`
	ID     string `json:"id"`
}

// Parent represents the parent of a Notion page.
type Parent struct {
	Type       string `json:"type"`
	DatabaseID string `json:"database_id,omitempty"`
	PageID     string `json:"page_id,omitempty"`
	BlockID    string `json:"block_id,omitempty"`
	Workspace  bool   `json:"workspace,omitempty"`
}

// Page represents a Notion page object.
type Page struct {
	Object         string                     `json:"object"`
	ID             string                     `json:"id"`
	CreatedTime    string                     `json:"created_time"`
	LastEditedTime string                     `json:"last_edited_time"`
	CreatedBy      PartialUser                `json:"created_by"`
	LastEditedBy   PartialUser                `json:"last_edited_by"`
	Parent         Parent                     `json:"parent"`
	InTrash        bool                       `json:"in_trash"`
	IsLocked       bool                       `json:"is_locked"`
	URL            string                     `json:"url"`
	PublicURL      *string                    `json:"public_url"`
	Icon           json.RawMessage            `json:"icon"`
	Cover          json.RawMessage            `json:"cover"`
	Properties     map[string]json.RawMessage `json:"properties"`
}

// CreatePageRequest is the body for creating a new page.
type CreatePageRequest struct {
	Parent     Parent         `json:"parent"`
	Properties map[string]any `json:"properties"`
	Children   []any          `json:"children,omitempty"`
	Icon       any            `json:"icon,omitempty"`
	Cover      any            `json:"cover,omitempty"`
}

// CreatePage creates a new page.
func (c *Client) CreatePage(ctx context.Context, req *CreatePageRequest) (*Page, error) {
	raw, err := c.Post(ctx, "/v1/pages", req)
	if err != nil {
		return nil, err
	}
	var p Page
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	return &p, nil
}

// UpdatePageRequest is the body for updating an existing page.
type UpdatePageRequest struct {
	Properties map[string]any `json:"properties,omitempty"`
	Archived   *bool          `json:"archived,omitempty"`
	Icon       any            `json:"icon,omitempty"`
	Cover      any            `json:"cover,omitempty"`
}

// UpdatePage updates an existing page by ID.
func (c *Client) UpdatePage(ctx context.Context, pageID string, req *UpdatePageRequest) (*Page, error) {
	raw, err := c.Patch(ctx, "/v1/pages/"+url.PathEscape(pageID), req)
	if err != nil {
		return nil, err
	}
	var p Page
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	return &p, nil
}

// GetPage retrieves a page by ID.
func (c *Client) GetPage(ctx context.Context, pageID string) (*Page, error) {
	raw, err := c.Get(ctx, "/v1/pages/"+url.PathEscape(pageID), nil)
	if err != nil {
		return nil, err
	}
	var p Page
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	return &p, nil
}
