package api

import (
	"context"
	"encoding/json"
	"net/url"
)

// pageResponse is a partial decode of a Notion list response.
type pageResponse struct {
	HasMore    bool    `json:"has_more"`
	NextCursor *string `json:"next_cursor"`
}

// Paginate calls fn for each page returned by the Notion list endpoint at path.
// It passes params on the first request and appends start_cursor on subsequent ones.
// Iteration stops when has_more is false or fn returns an error.
func Paginate(ctx context.Context, c *Client, path string, params url.Values, fn func(json.RawMessage) error) error {
	p := make(url.Values, len(params)+1)
	for k, v := range params {
		p[k] = v
	}

	for {
		raw, err := c.Get(ctx, path, p)
		if err != nil {
			return err
		}

		if err := fn(raw); err != nil {
			return err
		}

		var pr pageResponse
		if err := json.Unmarshal(raw, &pr); err != nil {
			return err
		}

		if !pr.HasMore || pr.NextCursor == nil {
			return nil
		}

		p.Set("start_cursor", *pr.NextCursor)
	}
}
