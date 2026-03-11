package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
