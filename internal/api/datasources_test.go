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

const dataSourceJSON = `{
	"object": "data_source",
	"id": "ds-1",
	"title": [{"type": "text", "text": {"content": "My Data Source"}, "plain_text": "My Data Source"}],
	"description": [],
	"parent": {"type": "database_id", "database_id": "db-1"},
	"database_parent": {"type": "page_id", "page_id": "page-1"},
	"is_inline": false,
	"in_trash": false,
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-2"},
	"properties": {},
	"icon": null,
	"cover": null,
	"url": "https://www.notion.so/ds-1",
	"public_url": null
}`

func TestListDataSources(t *testing.T) {
	t.Parallel()

	listResp := `{
		"object": "list",
		"results": [` + dataSourceJSON + `],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(listResp))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	sources, err := client.ListDataSources(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sources) != 1 {
		t.Fatalf("len = %d, want 1", len(sources))
	}
	if sources[0].ID != "ds-1" {
		t.Errorf("ID = %q, want %q", sources[0].ID, "ds-1")
	}
	if sources[0].Object != "data_source" {
		t.Errorf("Object = %q, want %q", sources[0].Object, "data_source")
	}
	if len(sources[0].Title) == 0 || sources[0].Title[0].PlainText != "My Data Source" {
		t.Errorf("Title plain_text = %q, want %q", sources[0].Title[0].PlainText, "My Data Source")
	}
}

func TestListDataSourcesPaginated(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)

		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_, _ = w.Write([]byte(`{
				"object": "list",
				"results": [` + dataSourceJSON + `],
				"has_more": true,
				"next_cursor": "cursor-2"
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"object": "list",
				"results": [` + dataSourceJSON + `],
				"has_more": false,
				"next_cursor": null
			}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	sources, err := client.ListDataSources(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if calls.Load() != 2 {
		t.Errorf("server calls = %d, want 2", calls.Load())
	}
	if len(sources) != 2 {
		t.Fatalf("len = %d, want 2", len(sources))
	}
}

func TestListDataSourcesError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"unauthorized","message":"API token is invalid."}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.ListDataSources(t.Context())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDataSourceDeserialization(t *testing.T) {
	t.Parallel()

	var ds api.DataSource
	if err := json.Unmarshal([]byte(dataSourceJSON), &ds); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ds.ID != "ds-1" {
		t.Errorf("ID = %q, want %q", ds.ID, "ds-1")
	}
	if ds.Object != "data_source" {
		t.Errorf("Object = %q, want %q", ds.Object, "data_source")
	}
	if ds.URL != "https://www.notion.so/ds-1" {
		t.Errorf("URL = %q, want %q", ds.URL, "https://www.notion.so/ds-1")
	}
	if ds.PublicURL != nil {
		t.Errorf("PublicURL = %v, want nil", ds.PublicURL)
	}
	if ds.InTrash {
		t.Error("InTrash = true, want false")
	}
	if ds.IsInline {
		t.Error("IsInline = true, want false")
	}
}

func TestCreateDataSource(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.CreateDataSourceRequest{
		Parent: api.Parent{Type: "database_id", DatabaseID: "db-1"},
	}
	ds, err := client.CreateDataSource(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds.ID != "ds-1" {
		t.Errorf("ID = %q, want %q", ds.ID, "ds-1")
	}
}

func TestCreateDataSourceError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"unauthorized","message":"API token is invalid."}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.CreateDataSource(t.Context(), &api.CreateDataSourceRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDataSource(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	ds, err := client.GetDataSource(t.Context(), "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds.ID != "ds-1" {
		t.Errorf("ID = %q, want %q", ds.ID, "ds-1")
	}
}

func TestGetDataSourceNotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"object_not_found","message":"not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetDataSource(t.Context(), "ds-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateDataSource(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1" || r.Method != http.MethodPatch {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.UpdateDataSourceRequest{}
	ds, err := client.UpdateDataSource(t.Context(), "ds-1", req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ds.ID != "ds-1" {
		t.Errorf("ID = %q, want %q", ds.ID, "ds-1")
	}
}

func TestUpdateDataSourceError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"unauthorized","message":"API token is invalid."}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.UpdateDataSource(t.Context(), "ds-1", &api.UpdateDataSourceRequest{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQueryDataSource(t *testing.T) {
	t.Parallel()

	pageJSON := `{"object":"page","id":"page-1"}`
	queryResp := `{
		"object": "list",
		"results": [` + pageJSON + `],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1/query" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(queryResp))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.QueryDataSource(t.Context(), "ds-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("len = %d, want 1", len(results))
	}
}

func TestQueryDataSourcePaginated(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	pageJSON := `{"object":"page","id":"page-1"}`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_, _ = w.Write([]byte(`{
				"object": "list",
				"results": [` + pageJSON + `],
				"has_more": true,
				"next_cursor": "cursor-2"
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"object": "list",
				"results": [` + pageJSON + `],
				"has_more": false,
				"next_cursor": null
			}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.QueryDataSource(t.Context(), "ds-1", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("server calls = %d, want 2", calls.Load())
	}
	if len(results) != 2 {
		t.Fatalf("len = %d, want 2", len(results))
	}
}

func TestQueryDataSourceError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"unauthorized","message":"API token is invalid."}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.QueryDataSource(t.Context(), "ds-1", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetDataSourceTemplates(t *testing.T) {
	t.Parallel()

	templatesResp := `{
		"templates": [
			{"id": "tmpl-1", "name": "Template One", "is_default": true},
			{"id": "tmpl-2", "name": "Template Two", "is_default": false}
		],
		"has_more": false,
		"next_cursor": null
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1/templates" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(templatesResp))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	templates, err := client.GetDataSourceTemplates(t.Context(), "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(templates) != 2 {
		t.Fatalf("len = %d, want 2", len(templates))
	}
	if templates[0].ID != "tmpl-1" {
		t.Errorf("ID = %q, want %q", templates[0].ID, "tmpl-1")
	}
	if !templates[0].IsDefault {
		t.Error("IsDefault = false, want true")
	}
	if templates[1].Name != "Template Two" {
		t.Errorf("Name = %q, want %q", templates[1].Name, "Template Two")
	}
}

func TestGetDataSourceTemplatesPaginated(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if n == 1 {
			_, _ = w.Write([]byte(`{
				"templates": [{"id": "tmpl-1", "name": "T1", "is_default": false}],
				"has_more": true,
				"next_cursor": "cursor-2"
			}`))
		} else {
			_, _ = w.Write([]byte(`{
				"templates": [{"id": "tmpl-2", "name": "T2", "is_default": false}],
				"has_more": false,
				"next_cursor": null
			}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	templates, err := client.GetDataSourceTemplates(t.Context(), "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls.Load() != 2 {
		t.Errorf("server calls = %d, want 2", calls.Load())
	}
	if len(templates) != 2 {
		t.Fatalf("len = %d, want 2", len(templates))
	}
}

func TestGetDataSourceTemplatesError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"code":"object_not_found","message":"not found"}`, http.StatusNotFound)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetDataSourceTemplates(t.Context(), "ds-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
