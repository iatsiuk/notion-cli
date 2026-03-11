package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/api"
)

const paragraphBlockJSON = `{
	"object": "block",
	"id": "block-1",
	"type": "paragraph",
	"has_children": false,
	"archived": false,
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-2"},
	"parent": {"type": "page_id", "page_id": "page-1"},
	"paragraph": {
		"rich_text": [{"type": "text", "text": {"content": "Hello"}, "plain_text": "Hello"}],
		"color": "default"
	}
}`

func TestGetBlock_ReturnsParagraph(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(paragraphBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	block, err := client.GetBlock(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if block.ID != "block-1" {
		t.Errorf("ID = %q, want %q", block.ID, "block-1")
	}
	if block.Object != "block" {
		t.Errorf("Object = %q, want %q", block.Object, "block")
	}
	if block.Type != "paragraph" {
		t.Errorf("Type = %q, want %q", block.Type, "paragraph")
	}
	if block.HasChildren {
		t.Error("HasChildren = true, want false")
	}
	if block.Archived {
		t.Error("Archived = true, want false")
	}
	if block.Parent.Type != "page_id" {
		t.Errorf("Parent.Type = %q, want %q", block.Parent.Type, "page_id")
	}
	if block.Parent.PageID != "page-1" {
		t.Errorf("Parent.PageID = %q, want %q", block.Parent.PageID, "page-1")
	}
	if block.Paragraph == nil {
		t.Error("Paragraph is nil, want non-nil")
	}
}

func TestGetBlock_Deserialization_Heading(t *testing.T) {
	t.Parallel()

	const j = `{
		"object": "block",
		"id": "block-2",
		"type": "heading_1",
		"has_children": false,
		"archived": false,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"heading_1": {
			"rich_text": [{"type": "text", "text": {"content": "Title"}, "plain_text": "Title"}],
			"color": "default",
			"is_toggleable": false
		}
	}`

	var block api.Block
	if err := json.Unmarshal([]byte(j), &block); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if block.Type != "heading_1" {
		t.Errorf("Type = %q, want %q", block.Type, "heading_1")
	}
	if block.Heading1 == nil {
		t.Error("Heading1 is nil, want non-nil")
	}
}

func TestGetBlock_Deserialization_Code(t *testing.T) {
	t.Parallel()

	const j = `{
		"object": "block",
		"id": "block-3",
		"type": "code",
		"has_children": false,
		"archived": false,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"code": {
			"rich_text": [{"type": "text", "text": {"content": "fmt.Println()"}, "plain_text": "fmt.Println()"}],
			"language": "go",
			"caption": []
		}
	}`

	var block api.Block
	if err := json.Unmarshal([]byte(j), &block); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if block.Type != "code" {
		t.Errorf("Type = %q, want %q", block.Type, "code")
	}
	if block.Code == nil {
		t.Error("Code is nil, want non-nil")
	}
}

func TestGetBlock_Deserialization_BulletedList(t *testing.T) {
	t.Parallel()

	const j = `{
		"object": "block",
		"id": "block-4",
		"type": "bulleted_list_item",
		"has_children": false,
		"archived": false,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"bulleted_list_item": {
			"rich_text": [{"type": "text", "text": {"content": "Item"}, "plain_text": "Item"}],
			"color": "default"
		}
	}`

	var block api.Block
	if err := json.Unmarshal([]byte(j), &block); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if block.Type != "bulleted_list_item" {
		t.Errorf("Type = %q, want %q", block.Type, "bulleted_list_item")
	}
	if block.BulletedListItem == nil {
		t.Error("BulletedListItem is nil, want non-nil")
	}
}

func TestGetBlock_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"block not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetBlock(t.Context(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
}

func TestGetBlock_AuthError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	_, err := client.GetBlock(t.Context(), "block-1")
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

func TestGetBlock_CallsCorrectEndpoint(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(paragraphBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetBlock(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/blocks/block-1" {
		t.Errorf("path = %q, want %q", gotPath, "/v1/blocks/block-1")
	}
}
