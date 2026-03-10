package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	defaultBaseURL = "https://api.notion.com"
	// NotionVersion is the Notion API version sent on every request.
	NotionVersion = "2022-06-28"
)

// Client is an HTTP client for the Notion API.
type Client struct {
	token   string
	baseURL string
	http    *http.Client
	verbose io.Writer // nil means quiet mode
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL overrides the default Notion API base URL.
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = u
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Client) {
		c.http = h
	}
}

// WithVerbose enables request/response logging to w.
// Pass os.Stderr for production use. Nil disables logging (same as omitting).
func WithVerbose(w io.Writer) Option {
	return func(c *Client) {
		c.verbose = w
	}
}

// NewClient creates a new Notion API client with the given Bearer token.
func NewClient(token string, opts ...Option) *Client {
	c := &Client{
		token:   token,
		baseURL: defaultBaseURL,
		http:    &http.Client{},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// Get sends a GET request to path with optional query params.
// Returns raw JSON bytes.
func (c *Client) Get(ctx context.Context, path string, params url.Values) (json.RawMessage, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.do(req)
}

// Post sends a POST request to path with body marshalled as JSON.
// body may be nil to send an empty body.
func (c *Client) Post(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.sendWithBody(ctx, http.MethodPost, path, body)
}

// Patch sends a PATCH request to path with body marshalled as JSON.
func (c *Client) Patch(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.sendWithBody(ctx, http.MethodPatch, path, body)
}

// Put sends a PUT request to path with body marshalled as JSON.
func (c *Client) Put(ctx context.Context, path string, body any) (json.RawMessage, error) {
	return c.sendWithBody(ctx, http.MethodPut, path, body)
}

// Delete sends a DELETE request to path.
func (c *Client) Delete(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.baseURL+path, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.do(req)
}

func (c *Client) sendWithBody(ctx context.Context, method, path string, body any) (json.RawMessage, error) {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &buf)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	return c.do(req)
}

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Notion-Version", NotionVersion)
	req.Header.Set("Content-Type", "application/json")
}

func (c *Client) do(req *http.Request) (json.RawMessage, error) {
	c.setHeaders(req)

	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &ConnectionError{cause: err}
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if c.verbose != nil {
		_, _ = fmt.Fprintf(c.verbose, "%s %s -> %d (%s)\n",
			req.Method, req.URL, resp.StatusCode, time.Since(start))
	}

	if resp.StatusCode >= 400 {
		return nil, parseAPIError(resp.StatusCode, body)
	}

	return body, nil
}
