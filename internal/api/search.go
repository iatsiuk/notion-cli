package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// SearchFilter restricts search results by object type.
type SearchFilter struct {
	Value    string `json:"value"`
	Property string `json:"property"`
}

// SearchSort controls the sort order of search results.
type SearchSort struct {
	Direction string `json:"direction"`
	Timestamp string `json:"timestamp"`
}

// SearchRequest is the body for POST /v1/search.
type SearchRequest struct {
	Query       string        `json:"query,omitempty"`
	Filter      *SearchFilter `json:"filter,omitempty"`
	Sort        *SearchSort   `json:"sort,omitempty"`
	StartCursor string        `json:"start_cursor,omitempty"`
}

// Search sends POST /v1/search and returns all matching results (auto-paginated).
// Each result is a raw JSON object (page or database), identified by the "object" field.
func (c *Client) Search(ctx context.Context, req *SearchRequest) ([]json.RawMessage, error) {
	var r SearchRequest
	if req != nil {
		r = *req
	}
	all := make([]json.RawMessage, 0)
	for {
		raw, err := c.Post(ctx, "/v1/search", r)
		if err != nil {
			return nil, err
		}
		var resp listResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return nil, fmt.Errorf("decode search response: %w", err)
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
