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

const testUserJSON = `{
	"object": "user",
	"id": "user-1",
	"type": "person",
	"name": "Alice",
	"person": {"email": "alice@example.com"}
}`

func TestRunUserMe_OutputsUser(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testUserJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserMe(context.Background(), client, &buf, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "user-1") {
		t.Errorf("output missing user ID, got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("output missing user name, got: %s", out)
	}
}

func TestRunUserMe_AuthError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserMe(context.Background(), client, &buf, "json")
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

func TestRunUserMe_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testUserJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserMe(context.Background(), client, &buf, "jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// jsonl output: one JSON object per line
	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl output is not valid JSON: %v, got: %s", err, line)
	}
}

const testUsersListJSON = `{
	"object": "list",
	"results": [
		{"object":"user","id":"user-1","type":"person","name":"Alice","person":{"email":"alice@example.com"}},
		{"object":"user","id":"user-2","type":"bot","name":"Bot","bot":{"owner":{"type":"workspace"},"workspace_name":"Acme"}}
	],
	"next_cursor": null,
	"has_more": false
}`

func TestRunUserList_OutputsUsers(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testUsersListJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserList(context.Background(), client, &buf, "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "user-1") {
		t.Errorf("output missing user-1, got: %s", out)
	}
	if !strings.Contains(out, "user-2") {
		t.Errorf("output missing user-2, got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("output missing Alice, got: %s", out)
	}
}

func TestRunUserList_Pagination(t *testing.T) {
	t.Parallel()
	page := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		page++
		if page == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"user-1","type":"person","name":"Alice"}],"next_cursor":"cursor-2","has_more":true}`))
		} else {
			_, _ = w.Write([]byte(`{"object":"list","results":[{"object":"user","id":"user-2","type":"person","name":"Bob"}],"next_cursor":null,"has_more":false}`))
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserList(context.Background(), client, &buf, "jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "user-1") {
		t.Errorf("output missing user-1, got: %s", out)
	}
	if !strings.Contains(out, "user-2") {
		t.Errorf("output missing user-2, got: %s", out)
	}
}

func TestRunUserList_OutputFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testUsersListJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserList(context.Background(), client, &buf, "jsonl")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// jsonl: each line is valid JSON
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %s", len(lines), buf.String())
	}
	for _, line := range lines {
		var obj map[string]any
		if err := json.Unmarshal([]byte(line), &obj); err != nil {
			t.Errorf("line is not valid JSON: %v, got: %s", err, line)
		}
	}
}

func TestRunUserGet_OutputsUser(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/user-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testUserJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserGet(context.Background(), client, &buf, "json", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "user-1") {
		t.Errorf("output missing user ID, got: %s", out)
	}
	if !strings.Contains(out, "Alice") {
		t.Errorf("output missing user name, got: %s", out)
	}
}

func TestRunUserGet_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewUserGetCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunUserGet_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"Could not find user with ID: bad-id."}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runUserGet(context.Background(), client, &buf, "json", "bad-id")
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
