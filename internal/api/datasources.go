package api

import (
	"context"
	"encoding/json"
	"fmt"
)

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
