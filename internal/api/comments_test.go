package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"notion-cli/internal/api"
)

const commentJSON = `{
	"object": "comment",
	"id": "comment-1",
	"parent": {"type": "page_id", "page_id": "page-1"},
	"discussion_id": "discussion-1",
	"rich_text": [{"type": "text", "text": {"content": "Hello"}, "plain_text": "Hello"}],
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"}
}`

func TestListComments_ReturnsPaginatedComments(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/comments" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Query().Get("block_id") != "block-1" {
			http.Error(w, "missing block_id", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + commentJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	comments, err := client.ListComments(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(comments) != 1 {
		t.Fatalf("len(comments) = %d, want 1", len(comments))
	}
	if comments[0].ID != "comment-1" {
		t.Errorf("ID = %q, want %q", comments[0].ID, "comment-1")
	}
	if comments[0].DiscussionID != "discussion-1" {
		t.Errorf("DiscussionID = %q, want %q", comments[0].DiscussionID, "discussion-1")
	}
}

func TestListComments_HandlesPagination(t *testing.T) {
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

	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)
		if r.URL.Query().Get("start_cursor") == "" {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + commentJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + comment2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	comments, err := client.ListComments(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount.Load() != 2 {
		t.Errorf("callCount = %d, want 2", callCount.Load())
	}
	if len(comments) != 2 {
		t.Fatalf("len(comments) = %d, want 2", len(comments))
	}
	if comments[0].ID != "comment-1" {
		t.Errorf("comments[0].ID = %q, want %q", comments[0].ID, "comment-1")
	}
	if comments[1].ID != "comment-2" {
		t.Errorf("comments[1].ID = %q, want %q", comments[1].ID, "comment-2")
	}
}

func TestComment_Deserialization(t *testing.T) {
	t.Parallel()

	var c api.Comment
	if err := json.Unmarshal([]byte(commentJSON), &c); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if c.ID != "comment-1" {
		t.Errorf("ID = %q, want %q", c.ID, "comment-1")
	}
	if c.Object != "comment" {
		t.Errorf("Object = %q, want %q", c.Object, "comment")
	}
	if c.Parent.Type != "page_id" {
		t.Errorf("Parent.Type = %q, want %q", c.Parent.Type, "page_id")
	}
	if c.Parent.PageID != "page-1" {
		t.Errorf("Parent.PageID = %q, want %q", c.Parent.PageID, "page-1")
	}
	if c.DiscussionID != "discussion-1" {
		t.Errorf("DiscussionID = %q, want %q", c.DiscussionID, "discussion-1")
	}
	if len(c.RichText) != 1 {
		t.Fatalf("len(RichText) = %d, want 1", len(c.RichText))
	}
	if c.CreatedTime != "2024-01-01T00:00:00.000Z" {
		t.Errorf("CreatedTime = %q, want %q", c.CreatedTime, "2024-01-01T00:00:00.000Z")
	}
	if c.CreatedBy.ID != "user-1" {
		t.Errorf("CreatedBy.ID = %q, want %q", c.CreatedBy.ID, "user-1")
	}
}

func TestCreateComment_OnPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/comments" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		parent, ok := body["parent"].(map[string]any)
		if !ok || parent["page_id"] != "page-1" {
			http.Error(w, "bad parent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreateCommentRequest{
		Parent:   &api.Parent{PageID: "page-1"},
		RichText: []api.RichTextItem{{Type: "text", Text: &api.RichTextText{Content: "Hello"}}},
	}
	cmt, err := client.CreateComment(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmt.ID != "comment-1" {
		t.Errorf("ID = %q, want %q", cmt.ID, "comment-1")
	}
}

func TestCreateComment_InDiscussion(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		if body["discussion_id"] != "discussion-1" {
			http.Error(w, "bad discussion_id", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(commentJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreateCommentRequest{
		DiscussionID: "discussion-1",
		RichText:     []api.RichTextItem{{Type: "text", Text: &api.RichTextText{Content: "Hello"}}},
	}
	cmt, err := client.CreateComment(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmt.ID != "comment-1" {
		t.Errorf("ID = %q, want %q", cmt.ID, "comment-1")
	}
}

func TestCreateComment_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"status":403,"code":"restricted_resource","message":"Insufficient permissions."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreateCommentRequest{
		Parent:   &api.Parent{PageID: "page-1"},
		RichText: []api.RichTextItem{{Type: "text", Text: &api.RichTextText{Content: "Hello"}}},
	}
	_, err := client.CreateComment(t.Context(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 403 {
		t.Errorf("Status = %d, want 403", apiErr.Status)
	}
}

func TestListComments_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.ListComments(t.Context(), "block-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 401 {
		t.Errorf("Status = %d, want 401", apiErr.Status)
	}
}
