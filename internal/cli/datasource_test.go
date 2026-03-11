package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notion-cli/internal/api"
)

const testDataSourceJSON = `{
	"object": "data_source",
	"id": "ds-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"title": [{"type": "text", "plain_text": "My DS"}],
	"description": [],
	"parent": {"type": "workspace", "workspace": true},
	"url": "https://www.notion.so/ds-1",
	"public_url": null,
	"is_inline": false,
	"in_trash": false,
	"icon": null,
	"cover": null,
	"properties": {}
}`

const dsListResponse = `{
	"object": "list",
	"results": [` + testDataSourceJSON + `],
	"has_more": false,
	"next_cursor": null
}`

// --- list ---

func TestRunDSList_OutputsDataSources(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dsListResponse))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSList(context.Background(), client, &buf, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ds-1") {
		t.Errorf("output missing data source ID, got: %s", buf.String())
	}
}

func TestRunDSList_Empty(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false,"next_cursor":null}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSList(context.Background(), client, &buf, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "[]") {
		t.Errorf("expected empty JSON array, got: %s", buf.String())
	}
}

func TestRunDSList_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSList(context.Background(), client, &buf, "json")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("expected exit code %d, got %d", ExitAuth, cliErr.Code)
	}
}

// --- get ---

func TestRunDSGet_OutputsDataSource(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSGet(context.Background(), client, &buf, "json", "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ds-1") {
		t.Errorf("output missing data source ID, got: %s", buf.String())
	}
}

func TestRunDSGet_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewDSGetCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunDSGet_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSGet(context.Background(), client, &buf, "json", "bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("expected exit code %d, got %d", ExitAPI, cliErr.Code)
	}
}

// --- create ---

func TestRunDSCreate_OutputsDataSource(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSCreate(context.Background(), client, &buf, "json", "page_id:page-1", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ds-1") {
		t.Errorf("output missing data source ID, got: %s", buf.String())
	}
}

func TestRunDSCreate_WithTitle(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSCreate(context.Background(), client, &buf, "json", "page_id:page-1", "My DS", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ds-1") {
		t.Errorf("output missing data source ID, got: %s", buf.String())
	}
}

func TestRunDSCreate_MissingParent(t *testing.T) {
	t.Parallel()
	cmd := NewDSCreateCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --parent flag, got nil")
	}
}

func TestRunDSCreate_InvalidParent(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := runDSCreate(context.Background(), nil, &buf, "json", "invalid-format", "", "")
	if err == nil {
		t.Fatal("expected error for invalid parent, got nil")
	}
}

func TestRunDSCreate_InvalidProperties(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := runDSCreate(context.Background(), nil, &buf, "json", "page_id:page-1", "", "not-json")
	if err == nil {
		t.Fatal("expected error for invalid properties JSON, got nil")
	}
}

func TestRunDSCreate_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSCreate(context.Background(), client, &buf, "json", "page_id:page-1", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("expected exit code %d, got %d", ExitAuth, cliErr.Code)
	}
}

// --- update ---

func TestRunDSUpdate_OutputsDataSource(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1" || r.Method != http.MethodPatch {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDataSourceJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSUpdate(context.Background(), client, &buf, "json", "ds-1", "New Title", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "ds-1") {
		t.Errorf("output missing data source ID, got: %s", buf.String())
	}
}

func TestRunDSUpdate_NoFlags(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := runDSUpdate(context.Background(), nil, &buf, "json", "ds-1", "", "", "")
	if err == nil {
		t.Fatal("expected error when no flags provided, got nil")
	}
	if !strings.Contains(err.Error(), "at least one of") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunDSUpdate_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewDSUpdateCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunDSUpdate_InvalidProperties(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := runDSUpdate(context.Background(), nil, &buf, "json", "ds-1", "", "", "not-json")
	if err == nil {
		t.Fatal("expected error for invalid properties JSON, got nil")
	}
}

func TestRunDSUpdate_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSUpdate(context.Background(), client, &buf, "json", "ds-1", "New Title", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("expected exit code %d, got %d", ExitAPI, cliErr.Code)
	}
}

// --- query ---

const dsQueryResultJSON = `{"object":"page","id":"result-1"}`

func dsQueryServer(t *testing.T, dsID string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/"+dsID+"/query" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"object":"list","results":[` + dsQueryResultJSON + `],"has_more":false,"next_cursor":null}`))
	}))
}

func TestRunDSQuery_OutputsResults(t *testing.T) {
	t.Parallel()
	srv := dsQueryServer(t, "ds-1")
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSQuery(context.Background(), client, &buf, "json", "ds-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "result-1") {
		t.Errorf("output missing result ID, got: %s", buf.String())
	}
}

func TestRunDSQuery_WithFilter(t *testing.T) {
	t.Parallel()
	srv := dsQueryServer(t, "ds-1")
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSQuery(context.Background(), client, &buf, "json", "ds-1", `{"property":"Status","select":{"equals":"Active"}}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "result-1") {
		t.Errorf("output missing result ID, got: %s", buf.String())
	}
}

func TestRunDSQuery_InvalidFilter(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	err := runDSQuery(context.Background(), nil, &buf, "json", "ds-1", "not-json")
	if err == nil {
		t.Fatal("expected error for invalid filter JSON, got nil")
	}
}

func TestRunDSQuery_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewDSQueryCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunDSQuery_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSQuery(context.Background(), client, &buf, "json", "bad-id", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("expected exit code %d, got %d", ExitAPI, cliErr.Code)
	}
}

// --- templates ---

const dsTemplatesResponse = `{
	"templates": [
		{"id": "tmpl-1", "name": "Template One", "is_default": true}
	],
	"has_more": false,
	"next_cursor": null
}`

func TestRunDSTemplates_OutputsTemplates(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/data_sources/ds-1/templates" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dsTemplatesResponse))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSTemplates(context.Background(), client, &buf, "json", "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "tmpl-1") {
		t.Errorf("output missing template ID, got: %s", buf.String())
	}
}

func TestRunDSTemplates_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(dsTemplatesResponse))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSTemplates(context.Background(), client, &buf, "jsonl", "ds-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d", len(lines))
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &obj); err != nil {
		t.Fatalf("line is not valid JSON: %v, got: %s", err, lines[0])
	}
	if obj["id"] != "tmpl-1" {
		t.Errorf("id = %v, want %q", obj["id"], "tmpl-1")
	}
}

func TestRunDSTemplates_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewDSTemplatesCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunDSTemplates_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runDSTemplates(context.Background(), client, &buf, "json", "bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("expected exit code %d, got %d", ExitAPI, cliErr.Code)
	}
}
