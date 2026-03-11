package api_test

import (
	"encoding/json"
	"io"
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

func TestUpdateBlock_UpdatesContent(t *testing.T) {
	t.Parallel()

	const updatedBlockJSON = `{
		"object": "block",
		"id": "block-1",
		"type": "paragraph",
		"has_children": false,
		"archived": false,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-03T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-2"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"paragraph": {
			"rich_text": [{"type": "text", "text": {"content": "Updated"}, "plain_text": "Updated"}],
			"color": "default"
		}
	}`

	var gotMethod, gotPath string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(updatedBlockJSON))
	}))
	defer srv.Close()

	req := &api.UpdateBlockRequest{
		TypeContent: map[string]any{
			"paragraph": map[string]any{
				"rich_text": []any{
					map[string]any{"type": "text", "text": map[string]any{"content": "Updated"}},
				},
			},
		},
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	block, err := client.UpdateBlock(t.Context(), "block-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want %q", gotMethod, http.MethodPatch)
	}
	if gotPath != "/v1/blocks/block-1" {
		t.Errorf("path = %q, want %q", gotPath, "/v1/blocks/block-1")
	}
	if block.ID != "block-1" {
		t.Errorf("ID = %q, want %q", block.ID, "block-1")
	}
	if block.Type != "paragraph" {
		t.Errorf("Type = %q, want %q", block.Type, "paragraph")
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(gotBody, &bodyMap); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if _, ok := bodyMap["paragraph"]; !ok {
		t.Error("body missing 'paragraph' key")
	}
}

func TestUpdateBlock_Archives(t *testing.T) {
	t.Parallel()

	const archivedBlockJSON = `{
		"object": "block",
		"id": "block-1",
		"type": "paragraph",
		"has_children": false,
		"archived": true,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-04T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-2"},
		"parent": {"type": "page_id", "page_id": "page-1"}
	}`

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(archivedBlockJSON))
	}))
	defer srv.Close()

	archived := true
	req := &api.UpdateBlockRequest{Archived: &archived}
	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	block, err := client.UpdateBlock(t.Context(), "block-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !block.Archived {
		t.Error("Archived = false, want true")
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(gotBody, &bodyMap); err != nil {
		t.Fatalf("body is not valid JSON: %v", err)
	}
	if v, ok := bodyMap["archived"]; !ok || v != true {
		t.Errorf("body archived = %v, want true", v)
	}
}

func TestUpdateBlock_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"block not found"}`))
	}))
	defer srv.Close()

	req := &api.UpdateBlockRequest{}
	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.UpdateBlock(t.Context(), "nonexistent", req)
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

func TestListBlockChildren_ReturnsChildren(t *testing.T) {
	t.Parallel()

	const childrenJSON = `{
		"object": "list",
		"results": [
			{
				"object": "block",
				"id": "child-1",
				"type": "paragraph",
				"has_children": false,
				"archived": false,
				"created_time": "2024-01-01T00:00:00.000Z",
				"last_edited_time": "2024-01-01T00:00:00.000Z",
				"created_by": {"object": "user", "id": "user-1"},
				"last_edited_by": {"object": "user", "id": "user-1"},
				"parent": {"type": "block_id", "block_id": "block-1"},
				"paragraph": {"rich_text": [], "color": "default"}
			},
			{
				"object": "block",
				"id": "child-2",
				"type": "heading_1",
				"has_children": false,
				"archived": false,
				"created_time": "2024-01-01T00:00:00.000Z",
				"last_edited_time": "2024-01-01T00:00:00.000Z",
				"created_by": {"object": "user", "id": "user-1"},
				"last_edited_by": {"object": "user", "id": "user-1"},
				"parent": {"type": "block_id", "block_id": "block-1"},
				"heading_1": {"rich_text": [], "color": "default", "is_toggleable": false}
			}
		],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1/children" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(childrenJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	blocks, err := client.ListBlockChildren(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("len = %d, want 2", len(blocks))
	}
	if blocks[0].ID != "child-1" {
		t.Errorf("blocks[0].ID = %q, want %q", blocks[0].ID, "child-1")
	}
	if blocks[1].ID != "child-2" {
		t.Errorf("blocks[1].ID = %q, want %q", blocks[1].ID, "child-2")
	}
}

func TestListBlockChildren_EmptyChildren(t *testing.T) {
	t.Parallel()

	const emptyJSON = `{
		"object": "list",
		"results": [],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(emptyJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	blocks, err := client.ListBlockChildren(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 0 {
		t.Errorf("len = %d, want 0", len(blocks))
	}
}

func TestListBlockChildren_Paginated(t *testing.T) {
	t.Parallel()

	page1 := `{
		"object": "list",
		"results": [{"object":"block","id":"child-1","type":"paragraph","has_children":false,"archived":false,"created_time":"2024-01-01T00:00:00.000Z","last_edited_time":"2024-01-01T00:00:00.000Z","created_by":{"object":"user","id":"u1"},"last_edited_by":{"object":"user","id":"u1"},"parent":{"type":"block_id","block_id":"block-1"},"paragraph":{"rich_text":[]}}],
		"has_more": true,
		"next_cursor": "cursor-abc"
	}`
	page2 := `{
		"object": "list",
		"results": [{"object":"block","id":"child-2","type":"paragraph","has_children":false,"archived":false,"created_time":"2024-01-01T00:00:00.000Z","last_edited_time":"2024-01-01T00:00:00.000Z","created_by":{"object":"user","id":"u1"},"last_edited_by":{"object":"user","id":"u1"},"parent":{"type":"block_id","block_id":"block-1"},"paragraph":{"rich_text":[]}}],
		"has_more": false,
		"next_cursor": null
	}`

	var requestCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("start_cursor") == "cursor-abc" {
			_, _ = w.Write([]byte(page2))
		} else {
			_, _ = w.Write([]byte(page1))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	blocks, err := client.ListBlockChildren(t.Context(), "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(blocks) != 2 {
		t.Fatalf("len = %d, want 2", len(blocks))
	}
	if requestCount != 2 {
		t.Errorf("requestCount = %d, want 2", requestCount)
	}
}

func TestListBlockChildren_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"block not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.ListBlockChildren(t.Context(), "nonexistent")
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
