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

func TestSearch_WithQuery(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Query string `json:"query"`
		}
		if err := json.Unmarshal(body, &payload); err != nil || payload.Query != "meeting notes" {
			http.Error(w, "bad query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + pageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.Search(t.Context(), &api.SearchRequest{Query: "meeting notes"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	var p api.Page
	if err := json.Unmarshal(results[0], &p); err != nil {
		t.Fatalf("unmarshal page: %v", err)
	}
	if p.ID != "page-1" {
		t.Errorf("ID = %q, want %q", p.ID, "page-1")
	}
}

func TestSearch_FilterPages(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if payload.Filter.Value != "page" || payload.Filter.Property != "object" {
			http.Error(w, "bad filter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + pageJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.SearchRequest{
		Filter: &api.SearchFilter{Value: "page", Property: "object"},
	}
	results, err := client.Search(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
}

func TestSearch_FilterDatabases(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		if payload.Filter.Value != "database" || payload.Filter.Property != "object" {
			http.Error(w, "bad filter", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + databaseJSON + `],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.SearchRequest{
		Filter: &api.SearchFilter{Value: "database", Property: "object"},
	}
	results, err := client.Search(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("want 1 result, got %d", len(results))
	}
	var db api.Database
	if err := json.Unmarshal(results[0], &db); err != nil {
		t.Fatalf("unmarshal database: %v", err)
	}
	if db.ID != "db-1" {
		t.Errorf("ID = %q, want %q", db.ID, "db-1")
	}
}

func TestSearch_WithSort(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload struct {
			Sort struct {
				Direction string `json:"direction"`
				Timestamp string `json:"timestamp"`
			} `json:"sort"`
		}
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		if payload.Sort.Direction != "descending" || payload.Sort.Timestamp != "last_edited_time" {
			http.Error(w, "bad sort", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	req := &api.SearchRequest{
		Sort: &api.SearchSort{Direction: "descending", Timestamp: "last_edited_time"},
	}
	results, err := client.Search(t.Context(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

func TestSearch_Pagination(t *testing.T) {
	t.Parallel()

	page2JSON := `{
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
			_, _ = w.Write([]byte(`{"object":"list","results":[` + pageJSON + `],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[` + page2JSON + `],"has_more":false,"next_cursor":null}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.Search(t.Context(), &api.SearchRequest{Query: "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("want 2 results, got %d", len(results))
	}

	var p1 api.Page
	if err := json.Unmarshal(results[0], &p1); err != nil {
		t.Fatalf("unmarshal page 1: %v", err)
	}
	if p1.ID != "page-1" {
		t.Errorf("results[0].ID = %q, want %q", p1.ID, "page-1")
	}

	var p2 api.Page
	if err := json.Unmarshal(results[1], &p2); err != nil {
		t.Fatalf("unmarshal page 2: %v", err)
	}
	if p2.ID != "page-2" {
		t.Errorf("results[1].ID = %q, want %q", p2.ID, "page-2")
	}

	if atomic.LoadInt32(&call) != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", atomic.LoadInt32(&call))
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.Search(t.Context(), &api.SearchRequest{Query: "nonexistent xyz"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}

func TestSearch_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.Search(t.Context(), &api.SearchRequest{Query: "test"})
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

func TestSearch_NilRequest(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	results, err := client.Search(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("want 0 results, got %d", len(results))
	}
}
