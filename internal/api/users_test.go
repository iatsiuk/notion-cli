package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"notion-cli/internal/api"
)

const userJSON = `{
	"object": "user",
	"id": "user-1",
	"type": "person",
	"name": "Alice",
	"avatar_url": "https://example.com/avatar.png",
	"person": {"email": "alice@example.com"}
}`

const botUserJSON = `{
	"object": "user",
	"id": "bot-1",
	"type": "bot",
	"name": "My Bot",
	"avatar_url": null,
	"bot": {
		"owner": {"type": "workspace", "workspace": true},
		"workspace_id": "ws-123",
		"workspace_limits": {"max_file_upload_size_in_bytes": 5242880},
		"workspace_name": "My Workspace"
	}
}`

func TestGetMe(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(userJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	user, err := client.GetMe(t.Context())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-1" {
		t.Errorf("ID = %q, want %q", user.ID, "user-1")
	}
	if user.Type != "person" {
		t.Errorf("Type = %q, want %q", user.Type, "person")
	}
	if user.Name == nil || *user.Name != "Alice" {
		t.Errorf("Name = %v, want %q", user.Name, "Alice")
	}
	if user.Person == nil || user.Person.Email != "alice@example.com" {
		t.Errorf("Person.Email = unexpected value, got %v", user.Person)
	}
}

func TestGetMe_AuthError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	_, err := client.GetMe(t.Context())
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

func TestListUsers(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		resp := map[string]any{
			"object":      "list",
			"results":     json.RawMessage(`[` + userJSON + `]`),
			"has_more":    false,
			"next_cursor": nil,
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	var users []api.User
	err := client.ListUsers(t.Context(), func(page []api.User) error {
		users = append(users, page...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 1 {
		t.Fatalf("want 1 user, got %d", len(users))
	}
	if users[0].ID != "user-1" {
		t.Errorf("ID = %q, want %q", users[0].ID, "user-1")
	}
}

func TestListUsers_Pagination(t *testing.T) {
	t.Parallel()

	var call int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := atomic.AddInt32(&call, 1)
		var resp map[string]any
		if c == 1 {
			resp = map[string]any{
				"object":      "list",
				"results":     json.RawMessage(`[` + userJSON + `]`),
				"has_more":    true,
				"next_cursor": "cursor1",
			}
		} else {
			resp = map[string]any{
				"object":      "list",
				"results":     json.RawMessage(`[` + botUserJSON + `]`),
				"has_more":    false,
				"next_cursor": nil,
			}
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	var users []api.User
	err := client.ListUsers(t.Context(), func(page []api.User) error {
		users = append(users, page...)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("want 2 users, got %d", len(users))
	}
}

func TestListUsers_APIError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"API token is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("bad-token", api.WithBaseURL(srv.URL))
	err := client.ListUsers(t.Context(), func(page []api.User) error { return nil })
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/user-1" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(userJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	user, err := client.GetUser(t.Context(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.ID != "user-1" {
		t.Errorf("ID = %q, want %q", user.ID, "user-1")
	}
	if user.Name == nil || *user.Name != "Alice" {
		t.Errorf("Name = %v, want %q", user.Name, "Alice")
	}
}

func TestGetUser_NotFound(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"status":404,"code":"object_not_found","message":"user not found"}`))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	_, err := client.GetUser(t.Context(), "nonexistent")
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

func TestGetUser_BotType(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(botUserJSON))
	}))
	defer srv.Close()

	client := api.NewClient("token", api.WithBaseURL(srv.URL))
	user, err := client.GetUser(t.Context(), "bot-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if user.Type != "bot" {
		t.Errorf("Type = %q, want %q", user.Type, "bot")
	}
	if user.Bot == nil || user.Bot.WorkspaceName == nil || *user.Bot.WorkspaceName != "My Workspace" {
		t.Errorf("Bot.WorkspaceName unexpected, got %v", user.Bot)
	}
	if user.Bot == nil || user.Bot.WorkspaceID != "ws-123" {
		t.Errorf("Bot.WorkspaceID = %q, want %q", user.Bot.WorkspaceID, "ws-123")
	}
	if user.Bot == nil || user.Bot.WorkspaceLimits.MaxFileUploadSizeInBytes != 5242880 {
		t.Errorf("Bot.WorkspaceLimits.MaxFileUploadSizeInBytes = %d, want %d", user.Bot.WorkspaceLimits.MaxFileUploadSizeInBytes, 5242880)
	}
	if user.AvatarURL != nil {
		t.Errorf("AvatarURL = %v, want nil", *user.AvatarURL)
	}
}
