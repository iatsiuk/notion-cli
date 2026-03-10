package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"notion-cli/internal/config"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()

	var buf bytes.Buffer
	done := make(chan error, 1)
	go func() {
		_, copyErr := io.Copy(&buf, r)
		done <- copyErr
	}()

	fn()
	_ = w.Close()

	if copyErr := <-done; copyErr != nil {
		t.Fatal(copyErr)
	}
	return buf.String()
}

func TestRunStatus_QuietSuppressesOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	cfg := &config.Config{Token: "test-token", Quiet: true}
	var out string
	var runErr error
	out = captureStdout(t, func() {
		runErr = runStatus(context.Background(), cfg, srv.URL, srv.Client())
	})
	if runErr != nil {
		t.Fatalf("expected no error, got: %v", runErr)
	}
	if out != "" {
		t.Errorf("expected no output with --quiet, got: %q", out)
	}
}

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
