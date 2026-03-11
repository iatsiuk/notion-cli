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

// propertyListPage is a partial decode of a paginated property response.
type propertyListPage struct {
	Object     string            `json:"object"`
	Type       string            `json:"type"`
	Results    []json.RawMessage `json:"results"`
	NextCursor *string           `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

// GetPageProperty retrieves a page property by ID.
// For paginated properties (title, rich_text, relation, rollup), collects all pages.
func (c *Client) GetPageProperty(ctx context.Context, pageID, propertyID string) (json.RawMessage, error) {
	path := "/v1/pages/" + url.PathEscape(pageID) + "/properties/" + url.PathEscape(propertyID)

	raw, err := c.Get(ctx, path, nil)
	if err != nil {
		return nil, err
	}

	var first propertyListPage
	if err := json.Unmarshal(raw, &first); err != nil {
		return nil, fmt.Errorf("decode property: %w", err)
	}

	if first.Object != "list" || !first.HasMore {
		return raw, nil
	}

	allResults := make([]json.RawMessage, 0, len(first.Results))
	allResults = append(allResults, first.Results...)

	cursor := first.NextCursor
	for cursor != nil {
		params := url.Values{"start_cursor": {*cursor}}
		raw, err = c.Get(ctx, path, params)
		if err != nil {
			return nil, err
		}
		var page propertyListPage
		if err := json.Unmarshal(raw, &page); err != nil {
			return nil, fmt.Errorf("decode property page: %w", err)
		}
		allResults = append(allResults, page.Results...)
		if !page.HasMore {
			cursor = nil
		} else {
			cursor = page.NextCursor
		}
	}

	merged := propertyListPage{
		Object:  first.Object,
		Type:    first.Type,
		Results: allResults,
		HasMore: false,
	}
	out, err := json.Marshal(merged)
	if err != nil {
		return nil, fmt.Errorf("encode merged property: %w", err)
	}
	return out, nil
}

// movePageRequest is the body for moving a page.
type movePageRequest struct {
	Parent Parent `json:"parent"`
}

// MovePage moves a page to a new parent.
func (c *Client) MovePage(ctx context.Context, pageID string, parent Parent) (*Page, error) {
	raw, err := c.Post(ctx, "/v1/pages/"+url.PathEscape(pageID)+"/move", &movePageRequest{Parent: parent})
	if err != nil {
		return nil, err
	}
	var p Page
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, fmt.Errorf("decode page: %w", err)
	}
	return &p, nil
}

// PageMarkdownResponse is the response from GET /v1/pages/{page_id}/markdown.
type PageMarkdownResponse struct {
	Object          string   `json:"object"`
	ID              string   `json:"id"`
	Markdown        string   `json:"markdown"`
	Truncated       bool     `json:"truncated"`
	UnknownBlockIDs []string `json:"unknown_block_ids"`
}

// GetPageMarkdown retrieves the markdown representation of a page.
func (c *Client) GetPageMarkdown(ctx context.Context, pageID string) (*PageMarkdownResponse, error) {
	raw, err := c.Get(ctx, "/v1/pages/"+url.PathEscape(pageID)+"/markdown", nil)
	if err != nil {
		return nil, err
	}
	var r PageMarkdownResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return nil, fmt.Errorf("decode page markdown: %w", err)
	}
	return &r, nil
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
