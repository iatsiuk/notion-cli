package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/api"
)

const fileUploadJSON = `{
	"object": "file_upload",
	"id": "fu-123",
	"created_time": "2024-01-01T00:00:00Z",
	"created_by": {"id": "user-1", "type": "person"},
	"last_edited_time": "2024-01-01T00:00:00Z",
	"in_trash": false,
	"expiry_time": null,
	"status": "pending",
	"filename": "test.pdf",
	"content_type": "application/pdf",
	"content_length": null,
	"upload_url": "https://upload.example.com/fu-123",
	"complete_url": "https://complete.example.com/fu-123",
	"number_of_parts": {"total": 1, "sent": 0}
}`

func TestCreateFileUpload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	fu, err := client.CreateFileUpload(t.Context(), api.CreateFileUploadParams{
		Filename:    "test.pdf",
		ContentType: "application/pdf",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if fu.ID != "fu-123" {
		t.Errorf("ID = %q, want %q", fu.ID, "fu-123")
	}
	if fu.Object != "file_upload" {
		t.Errorf("Object = %q, want %q", fu.Object, "file_upload")
	}
	if fu.Status != "pending" {
		t.Errorf("Status = %q, want %q", fu.Status, "pending")
	}
	if fu.Filename == nil || *fu.Filename != "test.pdf" {
		t.Errorf("Filename = %v, want %q", fu.Filename, "test.pdf")
	}
	if fu.ContentType == nil || *fu.ContentType != "application/pdf" {
		t.Errorf("ContentType = %v, want %q", fu.ContentType, "application/pdf")
	}
}

func TestCreateFileUpload_MinimalParams(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	// create with no optional params
	fu, err := client.CreateFileUpload(t.Context(), api.CreateFileUploadParams{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fu.ID != "fu-123" {
		t.Errorf("ID = %q, want %q", fu.ID, "fu-123")
	}
}

func TestCreateFileUpload_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	_, err := client.CreateFileUpload(t.Context(), api.CreateFileUploadParams{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 401 {
		t.Errorf("Status = %d, want 401", apiErr.Status)
	}
}

func TestGetFileUpload(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads/fu-123" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(fileUploadJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	fu, err := client.GetFileUpload(t.Context(), "fu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fu.ID != "fu-123" {
		t.Errorf("ID = %q, want %q", fu.ID, "fu-123")
	}
	if fu.Status != "pending" {
		t.Errorf("Status = %q, want %q", fu.Status, "pending")
	}
}

func TestGetFileUpload_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetFileUpload(t.Context(), "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 404 {
		t.Errorf("Status = %d, want 404", apiErr.Status)
	}
}

func TestDeleteFileUpload(t *testing.T) {
	t.Parallel()

	deletedJSON := `{
		"object": "file_upload",
		"id": "fu-123",
		"created_time": "2024-01-01T00:00:00Z",
		"created_by": {"id": "user-1", "type": "person"},
		"last_edited_time": "2024-01-01T00:00:00Z",
		"in_trash": true,
		"expiry_time": null,
		"status": "pending",
		"filename": "test.pdf",
		"content_type": "application/pdf",
		"content_length": null,
		"upload_url": "",
		"complete_url": "",
		"number_of_parts": {"total": 1, "sent": 0}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file_uploads/fu-123" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(deletedJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	fu, err := client.DeleteFileUpload(t.Context(), "fu-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fu.ID != "fu-123" {
		t.Errorf("ID = %q, want %q", fu.ID, "fu-123")
	}
	if !fu.InTrash {
		t.Error("expected InTrash = true")
	}
}

func TestDeleteFileUpload_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	_, err := client.DeleteFileUpload(t.Context(), "fu-123")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 401 {
		t.Errorf("Status = %d, want 401", apiErr.Status)
	}
}
