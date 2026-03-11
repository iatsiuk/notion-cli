package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// GetComment retrieves a comment by ID.
func (c *Client) GetComment(ctx context.Context, commentID string) (*Comment, error) {
	raw, err := c.Get(ctx, "/v1/comments/"+url.PathEscape(commentID), nil)
	if err != nil {
		return nil, err
	}
	var cmt Comment
	if err := json.Unmarshal(raw, &cmt); err != nil {
		return nil, fmt.Errorf("decode comment: %w", err)
	}
	return &cmt, nil
}

// DeleteComment deletes a comment by ID and returns the deleted comment.
func (c *Client) DeleteComment(ctx context.Context, commentID string) (*Comment, error) {
	raw, err := c.Delete(ctx, "/v1/comments/"+url.PathEscape(commentID))
	if err != nil {
		return nil, err
	}
	var cmt Comment
	if err := json.Unmarshal(raw, &cmt); err != nil {
		return nil, fmt.Errorf("decode comment: %w", err)
	}
	return &cmt, nil
}

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

// CreateCommentRequest is the body for creating a comment.
type CreateCommentRequest struct {
	Parent       *Parent        `json:"parent,omitempty"`
	DiscussionID string         `json:"discussion_id,omitempty"`
	RichText     []RichTextItem `json:"rich_text"`
}

// CreateComment creates a new comment on a page or in a discussion.
func (c *Client) CreateComment(ctx context.Context, req *CreateCommentRequest) (*Comment, error) {
	raw, err := c.Post(ctx, "/v1/comments", req)
	if err != nil {
		return nil, err
	}
	var cmt Comment
	if err := json.Unmarshal(raw, &cmt); err != nil {
		return nil, fmt.Errorf("decode comment: %w", err)
	}
	return &cmt, nil
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
