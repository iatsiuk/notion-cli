package cli

import (
	"bytes"
	"context"
	"errors"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"notion-cli/internal/api"
)

const testFileUploadJSON = `{
	"object": "file_upload",
	"id": "fu-1",
	"created_time": "2024-01-01T00:00:00.000Z",
	"created_by": {"id": "user-1", "type": "user"},
	"last_edited_time": "2024-01-01T00:00:00.000Z",
	"in_trash": false,
	"expiry_time": null,
	"status": "pending",
	"filename": "test.txt",
	"content_type": "text/plain",
	"content_length": 12,
	"upload_url": "https://upload.notion.so/fu-1",
	"complete_url": "https://api.notion.com/v1/file_uploads/fu-1/complete",
	"number_of_parts": {"total": 1, "sent": 0}
}`

func TestRunFileCreate_OutputsFileUpload(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileCreate(context.Background(), client, &buf, "json", "test.txt", "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fu-1") {
		t.Errorf("output missing file upload ID, got: %s", buf.String())
	}
}

func TestRunFileCreate_NoFilenameNoContentType(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileCreate(context.Background(), client, &buf, "json", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunFileCreate_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"invalid token"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileCreate(context.Background(), client, &buf, "json", "f.txt", "text/plain")
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

func TestRunFileGet_OutputsFileUpload(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads/fu-1" || r.Method != http.MethodGet {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileGet(context.Background(), client, &buf, "json", "fu-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fu-1") {
		t.Errorf("output missing file upload ID, got: %s", buf.String())
	}
}

func TestRunFileGet_NotFound(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileGet(context.Background(), client, &buf, "json", "bad-id")
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

func TestRunFileSend_SendsMultipart(t *testing.T) {
	t.Parallel()
	var gotContentType string
	var gotFilename string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads/fu-1/send" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		mediaType, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if !strings.HasPrefix(mediaType, "multipart/") {
			http.Error(w, "expected multipart", http.StatusBadRequest)
			return
		}
		mr := multipart.NewReader(r.Body, params["boundary"])
		p, err := mr.NextPart()
		if p == nil {
			http.Error(w, "expected multipart part, got nil: "+err.Error(), http.StatusBadRequest)
			return
		}
		gotContentType = p.Header.Get("Content-Type")
		_, params2, _ := mime.ParseMediaType(p.Header.Get("Content-Disposition"))
		gotFilename = params2["filename"]
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	// create a temp file
	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(fpath, []byte("hello world\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileSend(context.Background(), client, &buf, "json", "fu-1", fpath, "text/plain", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fu-1") {
		t.Errorf("output missing file upload ID, got: %s", buf.String())
	}
	if gotContentType != "text/plain" {
		t.Errorf("content-type = %q, want %q", gotContentType, "text/plain")
	}
	if gotFilename != "test.txt" {
		t.Errorf("filename = %q, want %q", gotFilename, "test.txt")
	}
}

func TestRunFileSend_FileNotFound(t *testing.T) {
	t.Parallel()
	client := api.NewClient("token")
	var buf bytes.Buffer
	err := runFileSend(context.Background(), client, &buf, "json", "fu-1", "/nonexistent/file.txt", "text/plain", 0)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestRunFileSend_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "test.txt")
	if err := os.WriteFile(fpath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileSend(context.Background(), client, &buf, "json", "bad-id", fpath, "text/plain", 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
}

func TestRunFileComplete_OutputsFileUpload(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads/fu-1/complete" || r.Method != http.MethodPost {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileComplete(context.Background(), client, &buf, "json", "fu-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fu-1") {
		t.Errorf("output missing file upload ID, got: %s", buf.String())
	}
}

func TestRunFileComplete_APIError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileComplete(context.Background(), client, &buf, "json", "bad-id")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *CLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected CLIError, got %T: %v", err, err)
	}
}

func TestNewFileCmd_HasSubcommands(t *testing.T) {
	t.Parallel()
	cmd := NewFileCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	for _, want := range []string{"create", "get", "send", "complete", "upload"} {
		if !names[want] {
			t.Errorf("missing subcommand %q", want)
		}
	}
}

func TestNewFileGetCmd_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewFileGetCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestNewFileCompleteCmd_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewFileCompleteCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}

func TestRunFileUpload_ChainsCreateSendComplete(t *testing.T) {
	t.Parallel()
	var calls []string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.Method+":"+r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(testFileUploadJSON))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(fpath, []byte("hello world\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileUpload(context.Background(), client, &buf, "json", fpath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "fu-1") {
		t.Errorf("output missing file upload ID, got: %s", buf.String())
	}
	mu.Lock()
	gotCalls := make([]string, len(calls))
	copy(gotCalls, calls)
	mu.Unlock()
	if len(gotCalls) != 3 {
		t.Errorf("expected 3 API calls, got %d: %v", len(gotCalls), gotCalls)
	}
	if gotCalls[0] != "POST:/v1/file_uploads" {
		t.Errorf("first call = %q, want POST:/v1/file_uploads", gotCalls[0])
	}
	if gotCalls[1] != "POST:/v1/file_uploads/fu-1/send" {
		t.Errorf("second call = %q, want POST:/v1/file_uploads/fu-1/send", gotCalls[1])
	}
	if gotCalls[2] != "POST:/v1/file_uploads/fu-1/complete" {
		t.Errorf("third call = %q, want POST:/v1/file_uploads/fu-1/complete", gotCalls[2])
	}
}

func TestRunFileUpload_FileNotFound(t *testing.T) {
	t.Parallel()
	client := api.NewClient("token")
	var buf bytes.Buffer
	err := runFileUpload(context.Background(), client, &buf, "json", "/nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestRunFileUpload_CreateError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"invalid token"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(fpath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileUpload(context.Background(), client, &buf, "json", fpath)
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

func TestRunFileUpload_SendError(t *testing.T) {
	t.Parallel()
	var callCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callCount.Add(1) == 1 {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(testFileUploadJSON))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	tmp := t.TempDir()
	fpath := filepath.Join(tmp, "hello.txt")
	if err := os.WriteFile(fpath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	client := api.NewClient("token", api.WithBaseURL(srv.URL), api.WithHTTPClient(srv.Client()))
	var buf bytes.Buffer
	err := runFileUpload(context.Background(), client, &buf, "json", fpath)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewFileCmd_HasUploadSubcommand(t *testing.T) {
	t.Parallel()
	cmd := NewFileCmd()
	names := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		names[sub.Name()] = true
	}
	if !names["upload"] {
		t.Error("missing subcommand \"upload\"")
	}
}

func TestNewFileUploadCmd_MissingArgument(t *testing.T) {
	t.Parallel()
	cmd := NewFileUploadCmd()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing argument, got nil")
	}
}
