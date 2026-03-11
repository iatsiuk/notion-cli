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

func TestRunOAuthToken_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/token" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "secret_abc",
			"token_type": "bearer",
			"bot_id": "bot-1",
			"workspace_id": "ws-1",
			"workspace_name": "Acme",
			"owner": {"type": "workspace", "workspace": true}
		}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthToken(context.Background(), client, &buf, "json", "cid", "csecret", "auth-code", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "secret_abc") {
		t.Errorf("output missing access_token, got: %s", out)
	}
	if !strings.Contains(out, "ws-1") {
		t.Errorf("output missing workspace_id, got: %s", out)
	}
}

func TestRunOAuthToken_JSONLFormat(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"access_token": "secret_abc",
			"token_type": "bearer",
			"bot_id": "bot-1",
			"workspace_id": "ws-1",
			"workspace_name": "Acme",
			"owner": {"type": "workspace", "workspace": true}
		}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthToken(context.Background(), client, &buf, "jsonl", "cid", "csecret", "auth-code", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := strings.TrimSpace(buf.String())
	var obj map[string]any
	if err := json.Unmarshal([]byte(line), &obj); err != nil {
		t.Fatalf("jsonl output not valid JSON: %v, got: %s", err, line)
	}
	if obj["access_token"] != "secret_abc" {
		t.Errorf("access_token = %v, want %q", obj["access_token"], "secret_abc")
	}
}

func TestRunOAuthToken_InvalidCode(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"invalid_grant","message":"The provided code is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthToken(context.Background(), client, &buf, "json", "cid", "csecret", "bad-code", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("exit code = %d, want %d", cliErr.Code, ExitAPI)
	}
}

func TestRunOAuthToken_Unauthorized(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"Client credentials are invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthToken(context.Background(), client, &buf, "json", "bad-id", "bad-secret", "code", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("exit code = %d, want %d", cliErr.Code, ExitAuth)
	}
}

func TestRunOAuthIntrospect_Active(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/introspect" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true,"scope":"read_content","iat":1700000000}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthIntrospect(context.Background(), client, &buf, "json", "cid", "csecret", "tok-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "read_content") {
		t.Errorf("output missing scope, got: %s", out)
	}
}

func TestRunOAuthIntrospect_Inactive(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":false}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthIntrospect(context.Background(), client, &buf, "json", "cid", "csecret", "expired")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var obj map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &obj); err != nil {
		t.Fatalf("output not valid JSON: %v, got: %s", err, buf.String())
	}
	if obj["active"] != false {
		t.Errorf("active = %v, want false", obj["active"])
	}
}

func TestRunOAuthIntrospect_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"Invalid client credentials."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthIntrospect(context.Background(), client, &buf, "json", "bad-cid", "bad-csecret", "tok")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("exit code = %d, want %d", cliErr.Code, ExitAuth)
	}
}

func TestRunOAuthRevoke_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/revoke" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_id":"req-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthRevoke(context.Background(), client, &buf, "cid", "csecret", "tok-to-revoke")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunOAuthRevoke_Error(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"invalid_request","message":"Token not found."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runOAuthRevoke(context.Background(), client, &buf, "cid", "csecret", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("exit code = %d, want %d", cliErr.Code, ExitAPI)
	}
}
