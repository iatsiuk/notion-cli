package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// RichTextItem is a single element of rich_text array.
type RichTextItem struct {
	Type      string          `json:"type"`
	Text      *RichTextText   `json:"text,omitempty"`
	PlainText string          `json:"plain_text,omitempty"`
	Href      *string         `json:"href,omitempty"`
	Mention   json.RawMessage `json:"mention,omitempty"`
	Equation  json.RawMessage `json:"equation,omitempty"`
}

// RichTextText is the text content of a rich_text item.
type RichTextText struct {
	Content string  `json:"content"`
	Link    *string `json:"link,omitempty"`
}

// Comment represents a Notion comment object.
type Comment struct {
	Object         string         `json:"object"`
	ID             string         `json:"id"`
	Parent         Parent         `json:"parent"`
	DiscussionID   string         `json:"discussion_id"`
	RichText       []RichTextItem `json:"rich_text"`
	CreatedTime    string         `json:"created_time"`
	LastEditedTime string         `json:"last_edited_time"`
	CreatedBy      PartialUser    `json:"created_by"`
}

// ListComments returns all comments for a block (auto-paginated).
func (c *Client) ListComments(ctx context.Context, blockID string) ([]Comment, error) {
	params := url.Values{"block_id": {blockID}}
	all := make([]Comment, 0)
	err := Paginate(ctx, c, "/v1/comments", params, func(raw json.RawMessage) error {
		var resp listResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("decode comments response: %w", err)
		}
		for _, r := range resp.Results {
			var cmt Comment
			if err := json.Unmarshal(r, &cmt); err != nil {
				return fmt.Errorf("decode comment: %w", err)
			}
			all = append(all, cmt)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return all, nil
}
