package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"notion-cli/internal/api"
)

const pageJSON = `{
	"object": "page",
	"id": "page-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-2"},
	"parent": {"type": "database_id", "database_id": "db-1"},
	"in_trash": false,
	"is_locked": false,
	"url": "https://www.notion.so/page-1",
	"public_url": null,
	"icon": {"type": "emoji", "emoji": "📄"},
	"cover": null,
	"properties": {
		"title": {
			"id": "title",
			"type": "title",
			"title": [{"type": "text", "text": {"content": "Hello"}, "plain_text": "Hello"}]
		}
	}
}`

func TestGetPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	page, err := client.GetPage(t.Context(), "page-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
	if page.Object != "page" {
		t.Errorf("Object = %q, want %q", page.Object, "page")
	}
	if page.Parent.Type != "database_id" {
		t.Errorf("Parent.Type = %q, want %q", page.Parent.Type, "database_id")
	}
	if page.Parent.DatabaseID != "db-1" {
		t.Errorf("Parent.DatabaseID = %q, want %q", page.Parent.DatabaseID, "db-1")
	}
	if page.InTrash {
		t.Error("InTrash = true, want false")
	}
	if page.URL != "https://www.notion.so/page-1" {
		t.Errorf("URL = %q, want %q", page.URL, "https://www.notion.so/page-1")
	}
	if _, ok := page.Properties["title"]; !ok {
		t.Error("Properties missing 'title'")
	}
}

func TestGetPage_Deserialization(t *testing.T) {
	t.Parallel()

	var page api.Page
	if err := json.Unmarshal([]byte(pageJSON), &page); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
	if page.CreatedTime != "2024-01-01T00:00:00.000Z" {
		t.Errorf("CreatedTime = %q, want %q", page.CreatedTime, "2024-01-01T00:00:00.000Z")
	}
	if page.CreatedBy.ID != "user-1" {
		t.Errorf("CreatedBy.ID = %q, want %q", page.CreatedBy.ID, "user-1")
	}
	if page.Parent.Type != "database_id" {
		t.Errorf("Parent.Type = %q, want %q", page.Parent.Type, "database_id")
	}
	if page.PublicURL != nil {
		t.Errorf("PublicURL = %v, want nil", *page.PublicURL)
	}
	if len(page.Properties) != 1 {
		t.Errorf("Properties len = %d, want 1", len(page.Properties))
	}
}

func TestGetPage_PageParent(t *testing.T) {
	t.Parallel()

	const j = `{
		"object": "page",
		"id": "page-2",
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-01T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "parent-page-1"},
		"in_trash": false,
		"is_locked": false,
		"url": "https://www.notion.so/page-2",
		"public_url": null,
		"icon": null,
		"cover": null,
		"properties": {}
	}`

	var page api.Page
	if err := json.Unmarshal([]byte(j), &page); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if page.Parent.Type != "page_id" {
		t.Errorf("Parent.Type = %q, want %q", page.Parent.Type, "page_id")
	}
	if page.Parent.PageID != "parent-page-1" {
		t.Errorf("Parent.PageID = %q, want %q", page.Parent.PageID, "parent-page-1")
	}
}

func TestGetPage_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"page not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetPage(t.Context(), "nonexistent")
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

func TestGetPage_AuthError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	_, err := client.GetPage(t.Context(), "page-1")
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

func TestCreatePage_WithDatabaseParent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		parent, _ := body["parent"].(map[string]any)
		if parent["database_id"] != "db-1" {
			http.Error(w, "wrong parent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreatePageRequest{
		Parent:     api.Parent{Type: "database_id", DatabaseID: "db-1"},
		Properties: map[string]any{"title": map[string]any{}},
	}
	page, err := client.CreatePage(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
}

func TestCreatePage_WithPageParent(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		parent, _ := body["parent"].(map[string]any)
		if parent["page_id"] != "parent-page-1" {
			http.Error(w, "wrong parent", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreatePageRequest{
		Parent:     api.Parent{Type: "page_id", PageID: "parent-page-1"},
		Properties: map[string]any{},
	}
	page, err := client.CreatePage(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
}

func TestCreatePage_SendsPropertiesPayload(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	props := map[string]any{
		"Name": map[string]any{"title": []any{map[string]any{"text": map[string]any{"content": "Test"}}}},
	}
	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreatePageRequest{
		Parent:     api.Parent{Type: "database_id", DatabaseID: "db-1"},
		Properties: props,
	}
	_, err := client.CreatePage(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bodyProps, ok := gotBody["properties"].(map[string]any)
	if !ok {
		t.Fatalf("properties missing in request body")
	}
	if _, ok := bodyProps["Name"]; !ok {
		t.Error("Name property missing in request body")
	}
}

func TestCreatePage_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"validation_error","message":"invalid parent"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreatePageRequest{
		Parent: api.Parent{Type: "database_id", DatabaseID: "bad"},
	}
	_, err := client.CreatePage(t.Context(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 400 {
		t.Errorf("Status = %d, want 400", apiErr.Status)
	}
}

func TestUpdatePage_UpdatesProperties(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1" || r.Method != http.MethodPatch {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	props := map[string]any{
		"Name": map[string]any{"title": []any{map[string]any{"text": map[string]any{"content": "Updated"}}}},
	}
	req := &api.UpdatePageRequest{Properties: props}
	page, err := client.UpdatePage(t.Context(), "page-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
	if _, ok := gotBody["properties"]; !ok {
		t.Error("properties missing in request body")
	}
}

func TestUpdatePage_Archives(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1" || r.Method != http.MethodPatch {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	archived := true
	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.UpdatePageRequest{Archived: &archived}
	_, err := client.UpdatePage(t.Context(), "page-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := gotBody["archived"].(bool)
	if !ok || !v {
		t.Errorf("archived = %v, want true", gotBody["archived"])
	}
}

func TestUpdatePage_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"page not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.UpdatePageRequest{}
	_, err := client.UpdatePage(t.Context(), "bad-id", req)
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

func TestGetPageProperty_ReturnsSingleItem(t *testing.T) {
	t.Parallel()

	const propJSON = `{"object":"property_item","type":"number","number":42,"id":"prop-1"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1/properties/prop-1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(propJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	raw, err := client.GetPageProperty(t.Context(), "page-1", "prop-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if obj["object"] != "property_item" {
		t.Errorf("object = %v, want %q", obj["object"], "property_item")
	}
	if obj["type"] != "number" {
		t.Errorf("type = %v, want %q", obj["type"], "number")
	}
}

func TestGetPageProperty_HandlesPagination(t *testing.T) {
	t.Parallel()

	const item1 = `{"object":"property_item","type":"rich_text","rich_text":{"plain_text":"Hello"}}`
	const item2 = `{"object":"property_item","type":"rich_text","rich_text":{"plain_text":"World"}}`

	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1/properties/prop-rt" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		callCount.Add(1)
		if r.URL.Query().Get("start_cursor") == "" {
			_, _ = w.Write([]byte(`{"object":"list","type":"property_item","results":[` + item1 + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","type":"property_item","results":[` + item2 + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	raw, err := client.GetPageProperty(t.Context(), "page-1", "prop-rt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount.Load() != 2 {
		t.Errorf("callCount = %d, want 2", callCount.Load())
	}

	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	results, ok := obj["results"].([]any)
	if !ok {
		t.Fatalf("results missing or wrong type")
	}
	if len(results) != 2 {
		t.Errorf("len(results) = %d, want 2", len(results))
	}
	if obj["has_more"] != false {
		t.Errorf("has_more = %v, want false", obj["has_more"])
	}
}

func TestGetPageProperty_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"property not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetPageProperty(t.Context(), "page-1", "bad-prop")
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

func TestMovePage_MovesToNewParent(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1/move" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(pageJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	page, err := client.MovePage(t.Context(), "page-1", api.Parent{Type: "page_id", PageID: "parent-2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page.ID != "page-1" {
		t.Errorf("ID = %q, want %q", page.ID, "page-1")
	}
	parent, ok := gotBody["parent"].(map[string]any)
	if !ok {
		t.Fatalf("parent missing in request body")
	}
	if parent["page_id"] != "parent-2" {
		t.Errorf("parent.page_id = %v, want %q", parent["page_id"], "parent-2")
	}
}

func TestMovePage_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"page not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.MovePage(t.Context(), "bad-id", api.Parent{Type: "page_id", PageID: "parent-1"})
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

func TestGetPageMarkdown_ReturnsMarkdown(t *testing.T) {
	t.Parallel()

	const resp = `{
		"object": "page_markdown",
		"id": "page-1",
		"markdown": "# Hello World\n\nSome content.",
		"truncated": false,
		"unknown_block_ids": []
	}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/pages/page-1/markdown" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(resp))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	result, err := client.GetPageMarkdown(t.Context(), "page-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Markdown != "# Hello World\n\nSome content." {
		t.Errorf("Markdown = %q, want %q", result.Markdown, "# Hello World\n\nSome content.")
	}
	if result.ID != "page-1" {
		t.Errorf("ID = %q, want %q", result.ID, "page-1")
	}
	if result.Truncated {
		t.Error("Truncated = true, want false")
	}
}

func TestGetPageMarkdown_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"page not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetPageMarkdown(t.Context(), "bad-id")
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
