package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestPaginate_SinglePage(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"results":     []any{"a", "b"},
			"has_more":    false,
			"next_cursor": nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("token", WithBaseURL(srv.URL))
	var pages []json.RawMessage
	err := Paginate(context.Background(), c, "/v1/list", url.Values{}, func(page json.RawMessage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 1 {
		t.Fatalf("want 1 page, got %d", len(pages))
	}
}

func TestPaginate_MultiplePages(t *testing.T) {
	t.Parallel()
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		var resp map[string]any
		if call == 1 {
			// first page - verify no start_cursor
			if r.URL.Query().Get("start_cursor") != "" {
				http.Error(w, "unexpected cursor", http.StatusBadRequest)
				return
			}
			resp = map[string]any{
				"results":     []any{"a"},
				"has_more":    true,
				"next_cursor": "cursor1",
			}
		} else {
			// second page - verify cursor forwarded
			if r.URL.Query().Get("start_cursor") != "cursor1" {
				http.Error(w, "missing cursor", http.StatusBadRequest)
				return
			}
			resp = map[string]any{
				"results":     []any{"b"},
				"has_more":    false,
				"next_cursor": nil,
			}
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("token", WithBaseURL(srv.URL))
	var pages []json.RawMessage
	err := Paginate(context.Background(), c, "/v1/list", url.Values{}, func(page json.RawMessage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("want 2 pages, got %d", len(pages))
	}
}

func TestPaginate_EmptyResults(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"results":     []any{},
			"has_more":    false,
			"next_cursor": nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("token", WithBaseURL(srv.URL))
	var pages []json.RawMessage
	err := Paginate(context.Background(), c, "/v1/list", url.Values{}, func(page json.RawMessage) error {
		pages = append(pages, page)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// handler is called once even for empty results
	if len(pages) != 1 {
		t.Fatalf("want 1 page, got %d", len(pages))
	}
}

func TestPaginate_ErrorMidPagination(t *testing.T) {
	t.Parallel()
	call := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		if call == 1 {
			resp := map[string]any{
				"results":     []any{"a"},
				"has_more":    true,
				"next_cursor": "cursor1",
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		// second page returns API error
		w.WriteHeader(http.StatusInternalServerError)
		resp := map[string]any{
			"object":  "error",
			"status":  500,
			"code":    "internal_server_error",
			"message": "server exploded",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("token", WithBaseURL(srv.URL))
	err := Paginate(context.Background(), c, "/v1/list", url.Values{}, func(page json.RawMessage) error {
		return nil
	})
	if err == nil {
		t.Fatal("want error, got nil")
	}
	var apiErr *APIError
	if !AsAPIError(err, &apiErr) {
		t.Fatalf("want *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 500 {
		t.Fatalf("want status 500, got %d", apiErr.Status)
	}
}

func TestPaginate_HandlerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"results":     []any{"a"},
			"has_more":    false,
			"next_cursor": nil,
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient("token", WithBaseURL(srv.URL))
	handlerErr := fmt.Errorf("handler failed")
	err := Paginate(context.Background(), c, "/v1/list", url.Values{}, func(page json.RawMessage) error {
		return handlerErr
	})
	if !errors.Is(err, handlerErr) {
		t.Fatalf("want handlerErr, got %v", err)
	}
}
