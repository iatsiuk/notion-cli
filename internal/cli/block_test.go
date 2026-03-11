package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"notion-cli/internal/api"
)

const testBlockJSON = `{
	"object": "block",
	"id": "block-1",
	"type": "paragraph",
	"has_children": false,
	"archived": false,
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"parent": {"type": "page_id", "page_id": "page-1"},
	"paragraph": {"rich_text": [], "color": "default"}
}`

func TestRunBlockGet_OutputsBlock(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockGet(context.Background(), client, &buf, "json", "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &obj); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if obj["id"] != "block-1" {
		t.Errorf("id = %v, want %q", obj["id"], "block-1")
	}
}

func TestRunBlockGet_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewBlockGetCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunBlockGet_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockGet(context.Background(), client, &buf, "jsonl", "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl output is not valid JSON: %v, got: %s", err, line)
	}
	if obj["id"] != "block-1" {
		t.Errorf("id = %v, want %q", obj["id"], "block-1")
	}
}

func TestRunBlockGet_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find block with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockGet(context.Background(), client, &buf, "json", "bad-id")
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

func TestRunBlockGet_AuthError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockGet(context.Background(), client, &buf, "json", "block-1")
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

func TestRunBlockUpdate_OutputsBlock(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1" || r.Method != http.MethodPatch {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "block-1", `{"paragraph":{"rich_text":[]}}`, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "block-1") {
		t.Errorf("output missing block ID, got: %s", out)
	}
}

func TestRunBlockUpdate_Archive(t *testing.T) {
	t.Parallel()

	const archivedJSON = `{
		"object": "block",
		"id": "block-1",
		"type": "paragraph",
		"has_children": false,
		"archived": true,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"paragraph": {"rich_text": [], "color": "default"}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(archivedJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "block-1", "{}", true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &obj); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if obj["archived"] != true {
		t.Errorf("archived = %v, want true", obj["archived"])
	}
}

func TestRunBlockUpdate_Unarchive(t *testing.T) {
	t.Parallel()

	const unarchivedJSON = `{
		"object": "block",
		"id": "block-1",
		"type": "paragraph",
		"has_children": false,
		"archived": false,
		"created_time": "2024-01-01T00:00:00.000Z",
		"last_edited_time": "2024-01-02T00:00:00.000Z",
		"created_by": {"object": "user", "id": "user-1"},
		"last_edited_by": {"object": "user", "id": "user-1"},
		"parent": {"type": "page_id", "page_id": "page-1"},
		"paragraph": {"rich_text": [], "color": "default"}
	}`

	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(unarchivedJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "block-1", "{}", false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var bodyMap map[string]any
	if err := json.Unmarshal(gotBody, &bodyMap); err != nil {
		t.Fatalf("request body is not valid JSON: %v", err)
	}
	if v, ok := bodyMap["archived"]; !ok || v != false {
		t.Errorf("body archived = %v, want false", v)
	}
}

func TestRunBlockUpdate_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "block-1", `not-json`, false, false)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestRunBlockUpdate_NoOp(t *testing.T) {
	t.Parallel()
	client := api.NewClient("token")
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "block-1", "{}", false, false)
	if err == nil {
		t.Fatal("expected error for no-op update, got nil")
	}
	if err.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
}

func TestRunBlockUpdate_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewBlockUpdateCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunBlockUpdate_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find block with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockUpdate(context.Background(), client, &buf, "json", "bad-id", "{}", true, false)
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

const testChildrenJSON = `{
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
		}
	],
	"has_more": false,
	"next_cursor": null
}`

func TestRunBlockChildren_OutputsChildren(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1/children" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testChildrenJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockChildren(context.Background(), client, &buf, "json", "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("len = %d, want 1", len(arr))
	}
	if obj, ok := arr[0].(map[string]any); !ok || obj["id"] != "child-1" {
		t.Errorf("arr[0].id = %v, want %q", arr[0], "child-1")
	}
}

func TestRunBlockChildren_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewBlockChildrenCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunBlockChildren_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find block with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockChildren(context.Background(), client, &buf, "json", "bad-id")
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

const testAppendResponseJSON = `{
	"object": "list",
	"results": [
		{
			"object": "block",
			"id": "new-child-1",
			"type": "paragraph",
			"has_children": false,
			"archived": false,
			"created_time": "2024-01-01T00:00:00.000Z",
			"last_edited_time": "2024-01-01T00:00:00.000Z",
			"created_by": {"object": "user", "id": "user-1"},
			"last_edited_by": {"object": "user", "id": "user-1"},
			"parent": {"type": "block_id", "block_id": "block-1"},
			"paragraph": {"rich_text": [], "color": "default"}
		}
	],
	"has_more": false,
	"next_cursor": null
}`

func TestRunBlockAppend_OutputsCreatedBlocks(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1/children" || r.Method != http.MethodPatch {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testAppendResponseJSON))
	}))
	defer srv.Close()

	childrenJSON := `[{"object":"block","type":"paragraph","paragraph":{"rich_text":[]}}]`
	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockAppend(context.Background(), client, &buf, "json", "block-1", childrenJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var arr []any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &arr); err != nil {
		t.Fatalf("output is not valid JSON array: %v", err)
	}
	if len(arr) != 1 {
		t.Fatalf("len = %d, want 1", len(arr))
	}
	if obj, ok := arr[0].(map[string]any); !ok || obj["id"] != "new-child-1" {
		t.Errorf("arr[0].id = %v, want %q", arr[0], "new-child-1")
	}
}

func TestRunBlockAppend_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewBlockAppendCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunBlockAppend_InvalidJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockAppend(context.Background(), client, &buf, "json", "block-1", `not-json`)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestRunBlockAppend_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find block with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockAppend(context.Background(), client, &buf, "json", "bad-id", `[]`)
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

const testDeletedBlockJSON = `{
	"object": "block",
	"id": "block-1",
	"type": "paragraph",
	"has_children": false,
	"archived": true,
	"created_time": "2024-01-01T00:00:00.000Z",
	"last_edited_time": "2024-01-02T00:00:00.000Z",
	"created_by": {"object": "user", "id": "user-1"},
	"last_edited_by": {"object": "user", "id": "user-1"},
	"parent": {"type": "page_id", "page_id": "page-1"},
	"paragraph": {"rich_text": [], "color": "default"}
}`

func TestRunBlockDelete_OutputsArchivedBlock(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/blocks/block-1" || r.Method != http.MethodDelete {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testDeletedBlockJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockDelete(context.Background(), client, &buf, "json", "block-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "block-1") {
		t.Errorf("output missing block ID, got: %s", out)
	}
	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &obj); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if obj["archived"] != true {
		t.Errorf("archived = %v, want true", obj["archived"])
	}
}

func TestRunBlockDelete_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewBlockDeleteCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunBlockDelete_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find block with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runBlockDelete(context.Background(), client, &buf, "json", "bad-id")
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
