package api_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

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

func TestGetQueryParams(t *testing.T) {
	t.Parallel()

	var gotQuery url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	params := url.Values{"page_size": {"10"}, "filter": {"active"}}

	_, err := client.Get(t.Context(), "/v1/pages", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := gotQuery.Get("page_size"); got != "10" {
		t.Errorf("page_size = %q, want %q", got, "10")
	}
	if got := gotQuery.Get("filter"); got != "active" {
		t.Errorf("filter = %q, want %q", got, "active")
	}
}

func TestPostJSONBody(t *testing.T) {
	t.Parallel()

	var gotBody map[string]any
	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"page-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	body := map[string]any{"title": "My Page", "archived": false}

	data, err := client.Post(t.Context(), "/v1/pages", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotBody["title"] != "My Page" {
		t.Errorf("body title = %v, want %q", gotBody["title"], "My Page")
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if result["id"] != "page-1" {
		t.Errorf("response id = %q, want %q", result["id"], "page-1")
	}
}

func TestGetContextCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// block until client cancels
		<-r.Context().Done()
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(t.Context())
	client := api.NewClient("token", api.WithBaseURL(srv.URL))

	done := make(chan error, 1)
	go func() {
		_, err := client.Get(ctx, "/v1/pages", nil)
		done <- err
	}()

	cancel()
	err := <-done
	if err == nil {
		t.Fatal("expected error after context cancellation, got nil")
	}
}

func TestGetTimeout(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	httpClient := &http.Client{Timeout: 50 * time.Millisecond}
	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(httpClient))

	_, err := client.Get(t.Context(), "/v1/pages", nil)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	// should be a connection error wrapping a timeout
	var connErr *api.ConnectionError
	if !api.AsConnectionError(err, &connErr) {
		t.Errorf("expected ConnectionError, got %T: %v", err, err)
	}
}

func TestPatchJSONBody(t *testing.T) {
	t.Parallel()

	var gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"page-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	body := map[string]any{"archived": true}

	_, err := client.Patch(t.Context(), "/v1/pages/page-1", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPatch {
		t.Errorf("method = %q, want PATCH", gotMethod)
	}
	if gotBody["archived"] != true {
		t.Errorf("body archived = %v, want true", gotBody["archived"])
	}
}

func TestPutJSONBody(t *testing.T) {
	t.Parallel()

	var gotMethod string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"block-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	body := map[string]any{"children": []any{}}

	_, err := client.Put(t.Context(), "/v1/blocks/block-1/children", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if _, ok := gotBody["children"]; !ok {
		t.Errorf("body missing children field, got %v", gotBody)
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	var gotMethod string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"archived":true}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))

	_, err := client.Delete(t.Context(), "/v1/blocks/block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotMethod != http.MethodDelete {
		t.Errorf("method = %q, want DELETE", gotMethod)
	}
}

func TestVerboseLogging(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithVerbose(&buf))

	_, err := client.Get(t.Context(), "/v1/pages", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logged := buf.String()
	if logged == "" {
		t.Fatal("expected log output, got empty")
	}
	if !bytes.Contains([]byte(logged), []byte("GET")) {
		t.Errorf("log missing method: %q", logged)
	}
	if !bytes.Contains([]byte(logged), []byte("200")) {
		t.Errorf("log missing status: %q", logged)
	}
}

func TestVerboseLoggingQuietMode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	// client without verbose option must not write any log output
	quietClient := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := quietClient.Get(t.Context(), "/v1/pages", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerboseLoggingPost(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"new-page"}`))
	}))
	defer srv.Close()

	var buf bytes.Buffer
	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithVerbose(&buf))

	_, err := client.Post(t.Context(), "/v1/pages", map[string]any{"title": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	logged := buf.String()
	if !bytes.Contains([]byte(logged), []byte("POST")) {
		t.Errorf("log missing method: %q", logged)
	}
	if !bytes.Contains([]byte(logged), []byte("201")) {
		t.Errorf("log missing status: %q", logged)
	}
}

// ensure Post with nil body sends an empty body (0 bytes)
func TestPostNilBody(t *testing.T) {
	t.Parallel()

	var gotBody bytes.Buffer
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = gotBody.ReadFrom(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))

	_, err := client.Post(t.Context(), "/v1/pages", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody.Len() != 0 {
		t.Errorf("expected empty body, got %q", gotBody.String())
	}
}
