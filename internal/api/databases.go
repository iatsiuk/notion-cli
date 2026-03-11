package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// RichText represents a Notion rich text object.
type RichText struct {
	Type      string          `json:"type"`
	Text      json.RawMessage `json:"text,omitempty"`
	PlainText string          `json:"plain_text,omitempty"`
	Href      *string         `json:"href,omitempty"`
}

// Database represents a Notion database object.
type Database struct {
	Object         string                     `json:"object"`
	ID             string                     `json:"id"`
	CreatedTime    string                     `json:"created_time"`
	LastEditedTime string                     `json:"last_edited_time"`
	CreatedBy      PartialUser                `json:"created_by"`
	LastEditedBy   PartialUser                `json:"last_edited_by"`
	Title          []RichText                 `json:"title"`
	Description    []RichText                 `json:"description"`
	Parent         Parent                     `json:"parent"`
	URL            string                     `json:"url"`
	PublicURL      *string                    `json:"public_url"`
	Archived       bool                       `json:"archived"`
	InTrash        bool                       `json:"in_trash"`
	IsInline       bool                       `json:"is_inline"`
	Icon           json.RawMessage            `json:"icon"`
	Cover          json.RawMessage            `json:"cover"`
	Properties     map[string]json.RawMessage `json:"properties"`
}

// searchRequest is the body for POST /v1/search.
type searchRequest struct {
	Filter      searchFilter `json:"filter"`
	StartCursor string       `json:"start_cursor,omitempty"`
}

type searchFilter struct {
	Value    string `json:"value"`
	Property string `json:"property"`
}

// searchListResponse is a partial decode of the paginated search response.
type searchListResponse struct {
	Results    []json.RawMessage `json:"results"`
	HasMore    bool              `json:"has_more"`
	NextCursor *string           `json:"next_cursor"`
}

// ListDatabases returns all databases accessible to the integration via the search API.
func (c *Client) ListDatabases(ctx context.Context) ([]Database, error) {
	req := searchRequest{
		Filter: searchFilter{Value: "database", Property: "object"},
	}
	all := make([]Database, 0)
	for {
		raw, err := c.Post(ctx, "/v1/search", req)
		if err != nil {
			return nil, err
		}
		var resp searchListResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode search response: %w", err)
		}
		for _, r := range resp.Results {
			var db Database
			if err := json.Unmarshal(r, &db); err != nil {
				return nil, fmt.Errorf("decode database: %w", err)
			}
			all = append(all, db)
		}
		if !resp.HasMore {
			return all, nil
		}
		if resp.NextCursor == nil {
			return nil, fmt.Errorf("pagination: has_more is true but next_cursor is nil")
		}
		req.StartCursor = *resp.NextCursor
	}
}

// CreateDatabaseRequest is the body for POST /v1/databases.
type CreateDatabaseRequest struct {
	Parent     Parent         `json:"parent"`
	Title      []any          `json:"title,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
}

// CreateDatabase creates a new database.
func (c *Client) CreateDatabase(ctx context.Context, req *CreateDatabaseRequest) (*Database, error) {
	raw, err := c.Post(ctx, "/v1/databases", req)
	if err != nil {
		return nil, err
	}
	var db Database
	if err := json.Unmarshal(raw, &db); err != nil {
		return nil, fmt.Errorf("decode database: %w", err)
	}
	return &db, nil
}

// UpdateDatabaseRequest is the body for PATCH /v1/databases/{id}.
type UpdateDatabaseRequest struct {
	Title       []any          `json:"title,omitempty"`
	Description []any          `json:"description,omitempty"`
	Properties  map[string]any `json:"properties,omitempty"`
}

// UpdateDatabase updates an existing database by ID.
func (c *Client) UpdateDatabase(ctx context.Context, databaseID string, req *UpdateDatabaseRequest) (*Database, error) {
	raw, err := c.Patch(ctx, "/v1/databases/"+url.PathEscape(databaseID), req)
	if err != nil {
		return nil, err
	}
	var db Database
	if err := json.Unmarshal(raw, &db); err != nil {
		return nil, fmt.Errorf("decode database: %w", err)
	}
	return &db, nil
}

// QueryDatabaseRequest is the body for POST /v1/databases/{id}/query.
type QueryDatabaseRequest struct {
	Filter      json.RawMessage `json:"filter,omitempty"`
	Sorts       json.RawMessage `json:"sorts,omitempty"`
	StartCursor string          `json:"start_cursor,omitempty"`
}

// queryDatabaseResponse is the paginated response from /v1/databases/{id}/query.
type queryDatabaseResponse struct {
	Results    []json.RawMessage `json:"results"`
	HasMore    bool              `json:"has_more"`
	NextCursor *string           `json:"next_cursor"`
}

// QueryDatabase queries a database and returns all matching pages (auto-paginated).
func (c *Client) QueryDatabase(ctx context.Context, databaseID string, req *QueryDatabaseRequest) ([]Page, error) {
	if req == nil {
		req = &QueryDatabaseRequest{}
	}
	path := "/v1/databases/" + url.PathEscape(databaseID) + "/query"
	all := make([]Page, 0)
	for {
		raw, err := c.Post(ctx, path, req)
		if err != nil {
			return nil, err
		}
		var resp queryDatabaseResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode query response: %w", err)
		}
		for _, r := range resp.Results {
			var p Page
			if err := json.Unmarshal(r, &p); err != nil {
				return nil, fmt.Errorf("decode page: %w", err)
			}
			all = append(all, p)
		}
		if !resp.HasMore {
			return all, nil
		}
		if resp.NextCursor == nil {
			return nil, fmt.Errorf("pagination: has_more is true but next_cursor is nil")
		}
		req.StartCursor = *resp.NextCursor
	}
}

// GetDatabase retrieves a database by ID.
func (c *Client) GetDatabase(ctx context.Context, databaseID string) (*Database, error) {
	raw, err := c.Get(ctx, "/v1/databases/"+url.PathEscape(databaseID), nil)
	if err != nil {
		return nil, err
	}
	var db Database
	if err := json.Unmarshal(raw, &db); err != nil {
		return nil, fmt.Errorf("decode database: %w", err)
	}
	return &db, nil
}
