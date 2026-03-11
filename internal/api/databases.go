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
