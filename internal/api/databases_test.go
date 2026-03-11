package api_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"notion-cli/internal/api"
)

const databaseJSON = `{
	"object": "database",
	"id": "db-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-2"},
	"title": [{"type": "text", "text": {"content": "My Database"}, "plain_text": "My Database"}],
	"description": [],
	"parent": {"type": "page_id", "page_id": "page-1"},
	"url": "https://www.notion.so/db-1",
	"public_url": null,
	"archived": false,
	"in_trash": false,
	"is_inline": false,
	"icon": null,
	"cover": null,
	"properties": {
		"Name": {
			"id": "title",
			"name": "Name",
			"type": "title",
			"title": {}
		},
		"Status": {
			"id": "status",
			"name": "Status",
			"type": "select",
			"select": {
				"options": [
					{"id": "opt-1", "name": "Active", "color": "green"}
				]
			}
		}
	}
}`

func TestGetDatabase(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(databaseJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	db, err := client.GetDatabase(t.Context(), "db-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
	if db.Object != "database" {
		t.Errorf("Object = %q, want %q", db.Object, "database")
	}
	if db.URL != "https://www.notion.so/db-1" {
		t.Errorf("URL = %q, want %q", db.URL, "https://www.notion.so/db-1")
	}
	if db.Parent.Type != "page_id" {
		t.Errorf("Parent.Type = %q, want %q", db.Parent.Type, "page_id")
	}
	if db.Parent.PageID != "page-1" {
		t.Errorf("Parent.PageID = %q, want %q", db.Parent.PageID, "page-1")
	}
	if db.Archived {
		t.Error("Archived = true, want false")
	}
}

func TestGetDatabase_Deserialization(t *testing.T) {
	t.Parallel()

	var db api.Database
	if err := json.Unmarshal([]byte(databaseJSON), &db); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
	if db.CreatedTime != "2024-01-01T00:00:00.000Z" {
		t.Errorf("CreatedTime = %q, want %q", db.CreatedTime, "2024-01-01T00:00:00.000Z")
	}
	if db.CreatedBy.ID != "user-1" {
		t.Errorf("CreatedBy.ID = %q, want %q", db.CreatedBy.ID, "user-1")
	}
	if db.PublicURL != nil {
		t.Errorf("PublicURL = %v, want nil", *db.PublicURL)
	}
	if len(db.Properties) != 2 {
		t.Errorf("Properties len = %d, want 2", len(db.Properties))
	}
	if _, ok := db.Properties["Name"]; !ok {
		t.Error("Properties missing 'Name'")
	}
	if _, ok := db.Properties["Status"]; !ok {
		t.Error("Properties missing 'Status'")
	}
	if len(db.Title) == 0 {
		t.Error("Title is empty, want at least one element")
	}
}

func TestListDatabases(t *testing.T) {
	t.Parallel()

	page1 := `{
		"object": "list",
		"results": [` + databaseJSON + `],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Filter struct {
				Value    string `json:"value"`
				Property string `json:"property"`
			} `json:"filter"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if payload.Filter.Value != "data_source" || payload.Filter.Property != "object" {
			http.Error(w, "bad filter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(page1))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	dbs, err := client.ListDatabases(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dbs) != 1 {
		t.Fatalf("want 1 database, got %d", len(dbs))
	}
	if dbs[0].ID != "db-1" {
		t.Errorf("ID = %q, want %q", dbs[0].ID, "db-1")
	}
}

func TestListDatabases_Pagination(t *testing.T) {
	t.Parallel()

	db2JSON := `{
		"object": "database",
		"id": "db-2",
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"title": [],
		"description": [],
		"parent": {"type": "workspace", "workspace": true},
		"url": "https://www.notion.so/db-2",
		"public_url": null,
		"archived": false,
		"in_trash": false,
		"is_inline": false,
		"icon": null,
		"cover": null,
		"properties": {}
	}`

	var call int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&call, 1)
		body, _ := io.ReadAll(r.Body)
		var req struct {
			StartCursor string `json:"start_cursor"`
		}
		_ = json.Unmarshal(body, &req)
		if c == 2 && req.StartCursor != "cursor1" {
			http.Error(w, "missing start_cursor", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if c == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + databaseJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + db2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	dbs, err := client.ListDatabases(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(dbs) != 2 {
		t.Fatalf("want 2 databases, got %d", len(dbs))
	}
	if dbs[0].ID != "db-1" {
		t.Errorf("dbs[0].ID = %q, want %q", dbs[0].ID, "db-1")
	}
	if dbs[1].ID != "db-2" {
		t.Errorf("dbs[1].ID = %q, want %q", dbs[1].ID, "db-2")
	}
	if atomic.LoadInt32(&call) != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", atomic.LoadInt32(&call))
	}
}

func TestListDatabases_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.ListDatabases(t.Context())
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

func TestGetDatabase_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"database not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetDatabase(t.Context(), "nonexistent")
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

func TestCreateDatabase(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(databaseJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreateDatabaseRequest{
		Parent: api.Parent{Type: "page_id", PageID: "page-1"},
	}
	db, err := client.CreateDatabase(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
	if db.Parent.PageID != "page-1" {
		t.Errorf("Parent.PageID = %q, want %q", db.Parent.PageID, "page-1")
	}
}

func TestCreateDatabase_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"validation_error","message":"parent is required"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.CreateDatabase(t.Context(), &api.CreateDatabaseRequest{})
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

func TestUpdateDatabase(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(databaseJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.UpdateDatabaseRequest{
		Title: []any{map[string]any{
			"type": "text",
			"text": map[string]any{"content": "Updated Title"},
		}},
	}
	db, err := client.UpdateDatabase(t.Context(), "db-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
}

func TestUpdateDatabase_Properties(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1" || r.Method != http.MethodPatch {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(databaseJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.UpdateDatabaseRequest{
		Properties: map[string]any{
			"Tags": map[string]any{"multi_select": map[string]any{}},
		},
	}
	db, err := client.UpdateDatabase(t.Context(), "db-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
}

func TestUpdateDatabase_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"database not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.UpdateDatabase(t.Context(), "bad-id", &api.UpdateDatabaseRequest{})
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

const queryPageJSON = `{
	"object": "page",
	"id": "page-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"parent": {"type": "database_id", "database_id": "db-1"},
	"url": "https://www.notion.so/page-1",
	"public_url": null,
	"in_trash": false,
	"is_locked": false,
	"icon": null,
	"cover": null,
	"properties": {}
}`

func TestQueryDatabase_NoFilter(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1/query" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + queryPageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	pages, err := client.QueryDatabase(t.Context(), "db-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("want 1 page, got %d", len(pages))
	}
	if pages[0].ID != "page-1" {
		t.Errorf("ID = %q, want %q", pages[0].ID, "page-1")
	}
}

func TestQueryDatabase_WithFilter(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1/query" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil || payload["filter"] == nil {
			http.Error(w, "missing filter in request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + queryPageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.QueryDatabaseRequest{
		Filter: []byte(`{"property":"Status","select":{"equals":"Active"}}`),
	}
	pages, err := client.QueryDatabase(t.Context(), "db-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("want 1 page, got %d", len(pages))
	}
}

func TestQueryDatabase_WithSorts(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/databases/db-1/query" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil || payload["sorts"] == nil {
			http.Error(w, "missing sorts in request body", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.QueryDatabaseRequest{
		Sorts: []byte(`[{"property":"Name","direction":"ascending"}]`),
	}
	pages, err := client.QueryDatabase(t.Context(), "db-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 0 {
		t.Fatalf("want 0 pages, got %d", len(pages))
	}
}

func TestQueryDatabase_Pagination(t *testing.T) {
	t.Parallel()

	page2JSON := `{
		"object": "page",
		"id": "page-2",
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "database_id", "database_id": "db-1"},
		"url": "https://www.notion.so/page-2",
		"public_url": null,
		"in_trash": false,
		"is_locked": false,
		"icon": null,
		"cover": null,
		"properties": {}
	}`

	var call int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&call, 1)
		body, _ := io.ReadAll(r.Body)
		var req struct {
			StartCursor string `json:"start_cursor"`
		}
		_ = json.Unmarshal(body, &req)
		if c == 2 && req.StartCursor != "cursor1" {
			http.Error(w, "missing start_cursor", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if c == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + queryPageJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + page2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	pages, err := client.QueryDatabase(t.Context(), "db-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
	if pages[0].ID != "page-1" {
		t.Errorf("pages[0].ID = %q, want %q", pages[0].ID, "page-1")
	}
	if pages[1].ID != "page-2" {
		t.Errorf("pages[1].ID = %q, want %q", pages[1].ID, "page-2")
	}
	if atomic.LoadInt32(&call) != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", atomic.LoadInt32(&call))
	}
}

func TestQueryDatabase_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"database not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.QueryDatabase(t.Context(), "bad-id", nil)
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
