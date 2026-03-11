package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UpdateBlockRequest is the body for updating a block.
// TypeContent maps block type names to their content (e.g. "paragraph" -> {...}).
type UpdateBlockRequest struct {
	InTrash     *bool          `json:"in_trash,omitempty"`
	TypeContent map[string]any `json:"-"`
}

// MarshalJSON merges TypeContent fields into the top-level JSON object.
func (r *UpdateBlockRequest) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	for k, v := range r.TypeContent {
		m[k] = v
	}
	if r.InTrash != nil {
		m["in_trash"] = *r.InTrash
	}
	return json.Marshal(m)
}

// Block represents a Notion block object.
type Block struct {
	Object         string      `json:"object"`
	ID             string      `json:"id"`
	Type           string      `json:"type"`
	HasChildren    bool        `json:"has_children"`
	InTrash        bool        `json:"in_trash"`
	CreatedTime    string      `json:"created_time"`
	LastEditedTime string      `json:"last_edited_time"`
	CreatedBy      PartialUser `json:"created_by"`
	LastEditedBy   PartialUser `json:"last_edited_by"`
	Parent         Parent      `json:"parent"`
	// type-specific content fields
	Paragraph        json.RawMessage `json:"paragraph,omitempty"`
	Heading1         json.RawMessage `json:"heading_1,omitempty"`
	Heading2         json.RawMessage `json:"heading_2,omitempty"`
	Heading3         json.RawMessage `json:"heading_3,omitempty"`
	BulletedListItem json.RawMessage `json:"bulleted_list_item,omitempty"`
	NumberedListItem json.RawMessage `json:"numbered_list_item,omitempty"`
	ToDo             json.RawMessage `json:"to_do,omitempty"`
	Toggle           json.RawMessage `json:"toggle,omitempty"`
	Code             json.RawMessage `json:"code,omitempty"`
	Image            json.RawMessage `json:"image,omitempty"`
	Quote            json.RawMessage `json:"quote,omitempty"`
	Callout          json.RawMessage `json:"callout,omitempty"`
	Divider          json.RawMessage `json:"divider,omitempty"`
	TableOfContents  json.RawMessage `json:"table_of_contents,omitempty"`
	Embed            json.RawMessage `json:"embed,omitempty"`
	Bookmark         json.RawMessage `json:"bookmark,omitempty"`
	Video            json.RawMessage `json:"video,omitempty"`
	File             json.RawMessage `json:"file,omitempty"`
	PDF              json.RawMessage `json:"pdf,omitempty"`
	Table            json.RawMessage `json:"table,omitempty"`
	TableRow         json.RawMessage `json:"table_row,omitempty"`
	ColumnList       json.RawMessage `json:"column_list,omitempty"`
	Column           json.RawMessage `json:"column,omitempty"`
	ChildPage        json.RawMessage `json:"child_page,omitempty"`
	ChildDatabase    json.RawMessage `json:"child_database,omitempty"`
	SyncedBlock      json.RawMessage `json:"synced_block,omitempty"`
	Template         json.RawMessage `json:"template,omitempty"`
	LinkPreview      json.RawMessage `json:"link_preview,omitempty"`
	Audio            json.RawMessage `json:"audio,omitempty"`
	Equation         json.RawMessage `json:"equation,omitempty"`
	Breadcrumb       json.RawMessage `json:"breadcrumb,omitempty"`
	LinkToPage       json.RawMessage `json:"link_to_page,omitempty"`
	MeetingNotes     json.RawMessage `json:"meeting_notes,omitempty"`
	Unsupported      json.RawMessage `json:"unsupported,omitempty"`
}

// UpdateBlock updates a block by ID.
func (c *Client) UpdateBlock(ctx context.Context, blockID string, req *UpdateBlockRequest) (*Block, error) {
	raw, err := c.Patch(ctx, "/v1/blocks/"+url.PathEscape(blockID), req)
	if err != nil {
		return nil, err
	}
	var b Block
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("decode block: %w", err)
	}
	return &b, nil
}

// ListBlockChildren returns all child blocks of a block (auto-paginated).
func (c *Client) ListBlockChildren(ctx context.Context, blockID string) ([]Block, error) {
	path := "/v1/blocks/" + url.PathEscape(blockID) + "/children"
	all := make([]Block, 0)
	err := Paginate(ctx, c, path, nil, func(raw json.RawMessage) error {
		var resp listResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("decode children response: %w", err)
		}
		for _, r := range resp.Results {
			var b Block
			if err := json.Unmarshal(r, &b); err != nil {
				return fmt.Errorf("decode block: %w", err)
			}
			all = append(all, b)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return all, nil
}

// AppendBlockChildren appends child blocks to a block and returns the created blocks.
func (c *Client) AppendBlockChildren(ctx context.Context, blockID string, children []map[string]any) ([]Block, error) {
	if len(children) == 0 {
		return nil, fmt.Errorf("children must be non-empty")
	}
	body := map[string]any{"children": children}
	raw, err := c.Patch(ctx, "/v1/blocks/"+url.PathEscape(blockID)+"/children", body)
	if err != nil {
		return nil, err
	}
	var resp listResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode append response: %w", err)
	}
	blocks := make([]Block, 0, len(resp.Results))
	for _, r := range resp.Results {
		var b Block
		if err := json.Unmarshal(r, &b); err != nil {
			return nil, fmt.Errorf("decode block: %w", err)
		}
		blocks = append(blocks, b)
	}
	return blocks, nil
}

// DeleteBlock moves a block to trash by ID and returns the updated block.
func (c *Client) DeleteBlock(ctx context.Context, blockID string) (*Block, error) {
	raw, err := c.Delete(ctx, "/v1/blocks/"+url.PathEscape(blockID))
	if err != nil {
		return nil, err
	}
	var b Block
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("decode block: %w", err)
	}
	return &b, nil
}

// GetBlock retrieves a block by ID.
func (c *Client) GetBlock(ctx context.Context, blockID string) (*Block, error) {
	raw, err := c.Get(ctx, "/v1/blocks/"+url.PathEscape(blockID), nil)
	if err != nil {
		return nil, err
	}
	var b Block
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("decode block: %w", err)
	}
	return &b, nil
}
