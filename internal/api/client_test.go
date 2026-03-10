package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/api"
)

func TestClientHeaders(t *testing.T) {
	t.Parallel()

	var gotReq *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := api.NewClient("secret-token", api.WithBaseURL(srv.URL))

	_, err := client.Get(t.Context(), "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := gotReq.Header.Get("Authorization"); got != "Bearer secret-token" {
		t.Errorf("Authorization = %q, want %q", got, "Bearer secret-token")
	}
	if got := gotReq.Header.Get("Notion-Version"); got != api.NotionVersion {
		t.Errorf("Notion-Version = %q, want %q", got, api.NotionVersion)
	}
	if got := gotReq.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestClientBaseURL(t *testing.T) {
	t.Parallel()

	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))

	_, err := client.Get(t.Context(), "/v1/pages/abc", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/v1/pages/abc" {
		t.Errorf("path = %q, want %q", gotPath, "/v1/pages/abc")
	}
}
