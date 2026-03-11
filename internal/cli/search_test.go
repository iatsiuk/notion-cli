package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notion-cli/internal/api"
)

const testSearchPageJSON = `{
	"object": "page",
	"id": "page-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"parent": {"type": "workspace", "workspace": true},
	"url": "https://www.notion.so/page-1",
	"public_url": null,
	"in_trash": false,
	"is_locked": false,
	"icon": null,
	"cover": null,
	"properties": {}
}`

const testSearchDBJSON = `{
	"object": "database",
	"id": "db-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"parent": {"type": "workspace", "workspace": true},
	"url": "https://www.notion.so/db-1",
	"public_url": null,
	"in_trash": false,
	"is_locked": false,
	"icon": null,
	"cover": null,
	"title": [],
	"description": [],
	"properties": {}
}`

func TestRunSearch_WithQuery(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body struct {
			Query string `json:"query"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Query != "meeting" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + testSearchPageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "meeting", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("len = %d, want 1", len(arr))
	}
	if obj, ok := arr[0].(map[string]any); !ok || obj["id"] != "page-1" {
		t.Errorf("arr[0].id = %v, want %q", arr[0], "page-1")
	}
}

func TestRunSearch_TypeFilterPage(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Filter struct {
				Value    string `json:"value"`
				Property string `json:"property"`
			} `json:"filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if body.Filter.Value != "page" || body.Filter.Property != "object" {
			http.Error(w, "bad filter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + testSearchPageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "test", "page", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("len = %d, want 1", len(arr))
	}
}

func TestRunSearch_TypeFilterDatabase(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Filter struct {
				Value string `json:"value"`
			} `json:"filter"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if body.Filter.Value != "database" {
			http.Error(w, "bad filter value", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + testSearchDBJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "test", "database", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("len = %d, want 1", len(arr))
	}
	if obj, ok := arr[0].(map[string]any); !ok || obj["id"] != "db-1" {
		t.Errorf("arr[0].id = %v, want %q", arr[0], "db-1")
	}
}

func TestRunSearch_SortDirection(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Sort struct {
				Direction string `json:"direction"`
				Timestamp string `json:"timestamp"`
			} `json:"sort"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if body.Sort.Direction != "descending" || body.Sort.Timestamp != "last_edited_time" {
			http.Error(w, "bad sort", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "test", "", "descending")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunSearch_HandlesPagination(t *testing.T) {
	t.Parallel()

	const page2JSON = `{
		"object": "page",
		"id": "page-2",
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "workspace", "workspace": true},
		"url": "https://www.notion.so/page-2",
		"public_url": null,
		"in_trash": false,
		"is_locked": false,
		"icon": null,
		"cover": null,
		"properties": {}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var body struct {
			StartCursor string `json:"start_cursor"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.StartCursor == "" {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + testSearchPageJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + page2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "test", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 2 {
		t.Fatalf("len = %d, want 2", len(arr))
	}
}

func TestRunSearch_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + testSearchPageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "jsonl", "test", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl output is not valid JSON: %v, got: %s", err, line)
	}
	if obj["id"] != "page-1" {
		t.Errorf("id = %v, want %q", obj["id"], "page-1")
	}
}

func TestNewSearchCmd_EmptyQuery(t *testing.T) {
	t.Parallel()
	cmd := NewSearchCmd()
	cmd.SetArgs([]string{""})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty query, got nil")
	}
	if err.Error() != "query must not be empty" {
		t.Errorf("unexpected error message: %q", err.Error())
	}
}

func TestRunSearch_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runSearch(context.Background(), client, &buf, "json", "test", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("expected exit code %d, got %d", ExitAuth, cliErr.Code)
	}
}
