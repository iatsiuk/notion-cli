package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// QueryDataSourceRequest is the body for POST /v1/data_sources/{id}/query.
type QueryDataSourceRequest struct {
	Filter      json.RawMessage `json:"filter,omitempty"`
	Sorts       json.RawMessage `json:"sorts,omitempty"`
	StartCursor string          `json:"start_cursor,omitempty"`
}

// DataSourceTemplate represents a template available in a data source.
type DataSourceTemplate struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsDefault bool   `json:"is_default"`
}

// templatesResponse is a partial decode of the GET /v1/data_sources/{id}/templates response.
type templatesResponse struct {
	Templates  []DataSourceTemplate `json:"templates"`
	HasMore    bool                 `json:"has_more"`
	NextCursor *string              `json:"next_cursor"`
}

// DataSource represents a Notion data source object.
type DataSource struct {
	Object         string                     `json:"object"`
	ID             string                     `json:"id"`
	Title          []RichText                 `json:"title"`
	Description    []RichText                 `json:"description"`
	Parent         json.RawMessage            `json:"parent"`
	DatabaseParent json.RawMessage            `json:"database_parent"`
	IsInline       bool                       `json:"is_inline"`
	InTrash        bool                       `json:"in_trash"`
	CreatedTime    string                     `json:"created_time"`
	LastEditedTime string                     `json:"last_edited_time"`
	CreatedBy      PartialUser                `json:"created_by"`
	LastEditedBy   PartialUser                `json:"last_edited_by"`
	Properties     map[string]json.RawMessage `json:"properties"`
	Icon           json.RawMessage            `json:"icon"`
	Cover          json.RawMessage            `json:"cover"`
	URL            string                     `json:"url"`
	PublicURL      *string                    `json:"public_url"`
}

// CreateDataSourceRequest is the body for POST /v1/data_sources.
type CreateDataSourceRequest struct {
	Parent     Parent          `json:"parent"`
	Title      []any           `json:"title,omitempty"`
	Properties *map[string]any `json:"properties,omitempty"`
}

// CreateDataSource creates a new data source.
func (c *Client) CreateDataSource(ctx context.Context, req *CreateDataSourceRequest) (*DataSource, error) {
	raw, err := c.Post(ctx, "/v1/data_sources", req)
	if err != nil {
		return nil, err
	}
	var ds DataSource
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, fmt.Errorf("decode data source: %w", err)
	}
	return &ds, nil
}

// GetDataSource retrieves a data source by ID.
func (c *Client) GetDataSource(ctx context.Context, dataSourceID string) (*DataSource, error) {
	raw, err := c.Get(ctx, "/v1/data_sources/"+url.PathEscape(dataSourceID), nil)
	if err != nil {
		return nil, err
	}
	var ds DataSource
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, fmt.Errorf("decode data source: %w", err)
	}
	return &ds, nil
}

// UpdateDataSourceRequest is the body for PATCH /v1/data_sources/{id}.
type UpdateDataSourceRequest struct {
	Title      []any           `json:"title,omitempty"`
	Properties *map[string]any `json:"properties,omitempty"`
}

// UpdateDataSource updates an existing data source by ID.
func (c *Client) UpdateDataSource(ctx context.Context, dataSourceID string, req *UpdateDataSourceRequest) (*DataSource, error) {
	raw, err := c.Patch(ctx, "/v1/data_sources/"+url.PathEscape(dataSourceID), req)
	if err != nil {
		return nil, err
	}
	var ds DataSource
	if err := json.Unmarshal(raw, &ds); err != nil {
		return nil, fmt.Errorf("decode data source: %w", err)
	}
	return &ds, nil
}

// QueryDataSource queries a data source and returns all matching results (auto-paginated).
func (c *Client) QueryDataSource(ctx context.Context, dataSourceID string, req *QueryDataSourceRequest) ([]json.RawMessage, error) {
	var r QueryDataSourceRequest
	if req != nil {
		r = *req
	}
	path := "/v1/data_sources/" + url.PathEscape(dataSourceID) + "/query"
	all := make([]json.RawMessage, 0)
	for {
		raw, err := c.Post(ctx, path, r)
		if err != nil {
			return nil, err
		}
		var resp listResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode query response: %w", err)
		}
		all = append(all, resp.Results...)
		if !resp.HasMore {
			return all, nil
		}
		if resp.NextCursor == nil {
			return nil, fmt.Errorf("pagination: has_more is true but next_cursor is nil")
		}
		r.StartCursor = *resp.NextCursor
	}
}

// GetDataSourceTemplates returns all templates for a data source (auto-paginated).
func (c *Client) GetDataSourceTemplates(ctx context.Context, dataSourceID string) ([]DataSourceTemplate, error) {
	path := "/v1/data_sources/" + url.PathEscape(dataSourceID) + "/templates"
	all := make([]DataSourceTemplate, 0)
	params := url.Values{}
	for {
		raw, err := c.Get(ctx, path, params)
		if err != nil {
			return nil, err
		}
		var resp templatesResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode templates response: %w", err)
		}
		all = append(all, resp.Templates...)
		if !resp.HasMore {
			return all, nil
		}
		if resp.NextCursor == nil {
			return nil, fmt.Errorf("pagination: has_more is true but next_cursor is nil")
		}
		params.Set("start_cursor", *resp.NextCursor)
	}
}

// ListDataSources returns all data sources accessible to the integration via the search API.
func (c *Client) ListDataSources(ctx context.Context) ([]DataSource, error) {
	req := SearchRequest{
		Filter: &SearchFilter{Value: "data_source", Property: "object"},
	}
	all := make([]DataSource, 0)
	for {
		raw, err := c.Post(ctx, "/v1/search", req)
		if err != nil {
			return nil, err
		}
		var resp listResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode search response: %w", err)
		}
		for _, r := range resp.Results {
			var ds DataSource
			if err := json.Unmarshal(r, &ds); err != nil {
				return nil, fmt.Errorf("decode data source: %w", err)
			}
			all = append(all, ds)
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
