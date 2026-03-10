package cli

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/config"
)

func TestRunStatus_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Notion-Version") == "" {
			t.Error("missing Notion-Version header")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"object":"user"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Token: "test-token"}
	err := runStatus(context.Background(), cfg, srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestRunStatus_AuthError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"object":"error","status":401,"code":"unauthorized"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Token: "bad-token"}
	err := runStatus(context.Background(), cfg, srv.URL, srv.Client())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("expected exit code %d, got %d", ExitAuth, cliErr.Code)
	}
}

func TestRunStatus_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"object":"error","status":500,"code":"internal_server_error"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Token: "test-token"}
	err := runStatus(context.Background(), cfg, srv.URL, srv.Client())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != ExitAPI {
		t.Errorf("expected exit code %d, got %d", ExitAPI, cliErr.Code)
	}
}

func TestRunStatus_ConnectionError(t *testing.T) {
	t.Parallel()
	cfg := &config.Config{Token: "test-token"}
	// use a closed server to simulate connection error
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	err := runStatus(context.Background(), cfg, srv.URL, srv.Client())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != ExitConnection {
		t.Errorf("expected exit code %d, got %d", ExitConnection, cliErr.Code)
	}
}

func TestRunStatus_ForbiddenIsAuth(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"object":"error","status":403,"code":"restricted_resource"}`))
	}))
	defer srv.Close()

	cfg := &config.Config{Token: "test-token"}
	err := runStatus(context.Background(), cfg, srv.URL, srv.Client())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T", err)
	}
	if cliErr.Code != ExitAuth {
		t.Errorf("expected exit code %d, got %d", ExitAuth, cliErr.Code)
	}
}
