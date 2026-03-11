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
	RefreshToken         *string    `json:"refresh_token,omitempty"` //nolint:gosec
	BotID                string     `json:"bot_id"`
	WorkspaceID          string     `json:"workspace_id"`
	WorkspaceName        *string    `json:"workspace_name"`
	WorkspaceIcon        *string    `json:"workspace_icon"`
	DuplicatedTemplateID *string    `json:"duplicated_template_id,omitempty"`
	RequestID            string     `json:"request_id,omitempty"`
	Owner                OAuthOwner `json:"owner"`
}

// OAuthRevokeResult is the response from POST /v1/oauth/revoke.
type OAuthRevokeResult struct {
	RequestID string `json:"request_id,omitempty"`
}

// tokenExchangeBody is the request body for POST /v1/oauth/token.
type tokenExchangeBody struct {
	GrantType   string `json:"grant_type"`
	Code        string `json:"code"`
	RedirectURI string `json:"redirect_uri,omitempty"`
}

// OAuthTokenInfo is the response from POST /v1/oauth/introspect.
type OAuthTokenInfo struct {
	Active    bool   `json:"active"`
	BotID     string `json:"bot_id,omitempty"`
	Scope     string `json:"scope,omitempty"`
	IAT       int64  `json:"iat,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// TokenExchange exchanges an authorization code for an access token.
// Uses HTTP Basic auth with clientID and clientSecret.
func (c *Client) TokenExchange(ctx context.Context, clientID, clientSecret, code, redirectURI string) (*OAuthToken, error) {
	body := tokenExchangeBody{
		GrantType:   "authorization_code",
		Code:        code,
		RedirectURI: redirectURI,
	}

	raw, err := c.oauthPost(ctx, clientID, clientSecret, "/v1/oauth/token", body)
	if err != nil {
		return nil, err
	}

	var tok OAuthToken
	if err := json.Unmarshal(raw, &tok); err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	return &tok, nil
}

// IntrospectToken introspects an access token.
// Uses HTTP Basic auth with clientID and clientSecret.
func (c *Client) IntrospectToken(ctx context.Context, clientID, clientSecret, token string) (*OAuthTokenInfo, error) {
	raw, err := c.oauthPost(ctx, clientID, clientSecret, "/v1/oauth/introspect", map[string]string{"token": token})
	if err != nil {
		return nil, err
	}

	var info OAuthTokenInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		return nil, fmt.Errorf("decode introspect: %w", err)
	}
	return &info, nil
}

// RevokeToken revokes an access token.
// Uses HTTP Basic auth with clientID and clientSecret.
func (c *Client) RevokeToken(ctx context.Context, clientID, clientSecret, token string) (*OAuthRevokeResult, error) {
	raw, err := c.oauthPost(ctx, clientID, clientSecret, "/v1/oauth/revoke", map[string]string{"token": token})
	if err != nil {
		return nil, err
	}

	var result OAuthRevokeResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("decode revoke: %w", err)
	}
	return &result, nil
}

// oauthPost sends a POST request with Basic auth and JSON body.
func (c *Client) oauthPost(ctx context.Context, clientID, clientSecret, path string, body interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("encode body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	creds := base64.StdEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	req.Header.Set("Authorization", "Basic "+creds)

	return c.do(req)
}
