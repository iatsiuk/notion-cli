package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

// OAuthOwner contains the owner info returned in a token response.
type OAuthOwner struct {
	Type      string `json:"type"`
	Workspace bool   `json:"workspace,omitempty"`
	User      *User  `json:"user,omitempty"`
}

// OAuthToken is the response from POST /v1/oauth/token.
type OAuthToken struct {
	AccessToken          string     `json:"access_token"` //nolint:gosec
	TokenType            string     `json:"token_type"`
	BotID                string     `json:"bot_id"`
	WorkspaceID          string     `json:"workspace_id"`
	WorkspaceName        string     `json:"workspace_name"`
	WorkspaceIcon        string     `json:"workspace_icon,omitempty"`
	DuplicatedTemplateID *string    `json:"duplicated_template_id,omitempty"`
	RequestID            string     `json:"request_id,omitempty"`
	Owner                OAuthOwner `json:"owner"`
}

// tokenExchangeBody is the request body for POST /v1/oauth/token.
type tokenExchangeBody struct {
	GrantType   string `json:"grant_type"`
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// TokenExchange exchanges an authorization code for an access token.
// Uses HTTP Basic auth with clientID and clientSecret.
func (c *Client) TokenExchange(ctx context.Context, clientID, clientSecret, code, redirectURI string) (*OAuthToken, error) {
	body := tokenExchangeBody{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: redirectURI,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/v1/oauth/token", &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)

	raw, err := c.do(req)
	if err != nil {
		return nil, err
	}

	var tok OAuthToken
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return &tok, nil
}
