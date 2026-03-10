package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// UserPerson contains person-specific user fields.
type UserPerson struct {
	Email string `json:"email"`
}

// BotOwner contains the owner info for a bot user.
type BotOwner struct {
	Type      string `json:"type"`
	Workspace bool   `json:"workspace,omitempty"`
	User      *User  `json:"user,omitempty"`
}

// BotWorkspaceLimits holds workspace limits for a bot user.
type BotWorkspaceLimits struct {
	MaxFileUploadSizeInBytes int `json:"max_file_upload_size_in_bytes"`
}

// UserBot contains bot-specific user fields.
type UserBot struct {
	Owner           BotOwner           `json:"owner"`
	WorkspaceID     string             `json:"workspace_id"`
	WorkspaceLimits BotWorkspaceLimits `json:"workspace_limits"`
	WorkspaceName   *string            `json:"workspace_name,omitempty"`
}

// User represents a Notion user object.
type User struct {
	ID        string      `json:"id"`
	Object    string      `json:"object"`
	Type      string      `json:"type"`
	Name      *string     `json:"name"`
	AvatarURL *string     `json:"avatar_url"`
	Person    *UserPerson `json:"person,omitempty"`
	Bot       *UserBot    `json:"bot,omitempty"`
}

// GetMe returns the currently authenticated user.
func (c *Client) GetMe(ctx context.Context) (*User, error) {
	raw, err := c.Get(ctx, "/v1/users/me", nil)
	if err != nil {
		return nil, err
	}
	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &u, nil
}

// usersResponse is a partial decode of the list users response.
type usersResponse struct {
	Results []User `json:"results"`
}

// ListUsers iterates over all workspace users, calling fn for each page.
func (c *Client) ListUsers(ctx context.Context, fn func([]User) error) error {
	return Paginate(ctx, c, "/v1/users", url.Values{}, func(raw json.RawMessage) error {
		var resp usersResponse
		if err := json.Unmarshal(raw, &resp); err != nil {
			return fmt.Errorf("decode users page: %w", err)
		}
		return fn(resp.Results)
	})
}

// GetUser returns a specific user by ID.
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	raw, err := c.Get(ctx, "/v1/users/"+url.PathEscape(userID), nil)
	if err != nil {
		return nil, err
	}
	var u User
	if err := json.Unmarshal(raw, &u); err != nil {
		return nil, fmt.Errorf("decode user: %w", err)
	}
	return &u, nil
}
