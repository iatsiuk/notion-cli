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

const testCommentJSON = `{
	"object": "comment",
	"id": "comment-1",
	"parent": {"type": "page_id", "page_id": "page-1"},
	"discussion_id": "discussion-1",
	"rich_text": [{"type": "text", "text": {"content": "Hello"}, "plain_text": "Hello"}],
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"}
}`

const testCommentListJSON = `{
	"object": "list",
	"results": [` + testCommentJSON + `],
	"has_more": false,
	"next_cursor": null
}`

func TestRunCommentList_OutputsComments(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/comments" || r.URL.Query().Get("block_id") != "block-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testCommentListJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runCommentList(context.Background(), client, &buf, "json", "block-1")
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
	if obj, ok := arr[0].(map[string]any); !ok || obj["id"] != "comment-1" {
		t.Errorf("arr[0].id = %v, want %q", arr[0], "comment-1")
	}
}

func TestRunCommentList_MissingBlockFlag(t *testing.T) {
	t.Parallel()
	cmd := NewCommentCmd()
	cmd.SetArgs([]string{"list"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --block flag, got nil")
	}
}

func TestRunCommentList_HandlesPagination(t *testing.T) {
	t.Parallel()

	const comment2JSON = `{
		"object": "comment",
		"id": "comment-2",
		"parent": {"type": "page_id", "page_id": "page-1"},
		"discussion_id": "discussion-1",
		"rich_text": [],
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("start_cursor") == "" {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + testCommentJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + comment2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runCommentList(context.Background(), client, &buf, "json", "block-1")
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

func TestRunCommentList_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testCommentListJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runCommentList(context.Background(), client, &buf, "jsonl", "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl output is not valid JSON: %v, got: %s", err, line)
	}
	if obj["id"] != "comment-1" {
		t.Errorf("id = %v, want %q", obj["id"], "comment-1")
	}
}

func TestRunCommentList_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runCommentList(context.Background(), client, &buf, "json", "block-1")
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
