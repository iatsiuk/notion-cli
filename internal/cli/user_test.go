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
