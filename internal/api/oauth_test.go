package api_test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/api"
)

const oauthTokenJSON = `{
	"access_token": "secret_abc123",
	"token_type": "bearer",
	"bot_id": "bot-1",
	"workspace_id": "ws-1",
	"workspace_name": "Acme",
	"workspace_icon": "https://example.com/icon.png",
	"owner": {"type": "workspace", "workspace": true}
}`

func TestTokenExchange_Payload(t *testing.T) {
	t.Parallel()

	type capture struct {
		auth        string
		grantType   string
		code        string
		redirectURI string
	}
	ch := make(chan capture, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/token" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ch <- capture{
			auth:        r.Header.Get("Authorization"),
			grantType:   body["grant_type"],
			code:        body["code"],
			redirectURI: body["redirect_uri"],
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(oauthTokenJSON))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	tok, err := client.TokenExchange(t.Context(), "client-id", "client-secret", "auth-code", "https://example.com/cb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := <-ch

	// verify Basic auth header
	expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("client-id:client-secret"))
	if got.auth != expected {
		t.Errorf("Authorization = %q, want %q", got.auth, expected)
	}
	if got.grantType != "authorization_code" {
		t.Errorf("grant_type = %q, want %q", got.grantType, "authorization_code")
	}
	if got.code != "auth-code" {
		t.Errorf("code = %q, want %q", got.code, "auth-code")
	}
	if got.redirectURI != "https://example.com/cb" {
		t.Errorf("redirect_uri = %q, want %q", got.redirectURI, "https://example.com/cb")
	}

	// verify response fields
	if tok.AccessToken != "secret_abc123" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "secret_abc123")
	}
	if tok.TokenType != "bearer" {
		t.Errorf("TokenType = %q, want %q", tok.TokenType, "bearer")
	}
	if tok.BotID != "bot-1" {
		t.Errorf("BotID = %q, want %q", tok.BotID, "bot-1")
	}
	if tok.WorkspaceID != "ws-1" {
		t.Errorf("WorkspaceID = %q, want %q", tok.WorkspaceID, "ws-1")
	}
	if tok.WorkspaceName == nil || *tok.WorkspaceName != "Acme" {
		got := "<nil>"
		if tok.WorkspaceName != nil {
			got = *tok.WorkspaceName
		}
		t.Errorf("WorkspaceName = %q, want %q", got, "Acme")
	}
}

func TestTokenExchange_InvalidCode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"invalid_grant","message":"The provided code is invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	_, err := client.TokenExchange(t.Context(), "client-id", "client-secret", "bad-code", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Status != 400 {
		t.Errorf("Status = %d, want 400", apiErr.Status)
	}
}

func TestTokenExchange_ExpiredCode(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"invalid_grant","message":"The authorization code has expired."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	_, err := client.TokenExchange(t.Context(), "client-id", "client-secret", "expired-code", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "invalid_grant" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "invalid_grant")
	}
}

func TestTokenExchange_Unauthorized(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"Client credentials are invalid."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	_, err := client.TokenExchange(t.Context(), "bad-id", "bad-secret", "code", "")
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

func TestIntrospectToken_Active(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/introspect" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// verify Basic auth
		authHdr := r.Header.Get("Authorization")
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("cid:csecret"))
		if authHdr != expected {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["token"] == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":true,"scope":"read_content","iat":1700000000,"request_id":"req-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	info, err := client.IntrospectToken(t.Context(), "cid", "csecret", "tok-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !info.Active {
		t.Error("Active = false, want true")
	}
	if info.Scope != "read_content" {
		t.Errorf("Scope = %q, want %q", info.Scope, "read_content")
	}
	if info.IAT != 1700000000 {
		t.Errorf("IAT = %d, want 1700000000", info.IAT)
	}
}

func TestIntrospectToken_Inactive(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"active":false}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	info, err := client.IntrospectToken(t.Context(), "cid", "csecret", "expired-tok")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Active {
		t.Error("Active = true, want false")
	}
}

func TestIntrospectToken_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"status":401,"code":"unauthorized","message":"Invalid client credentials."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	_, err := client.IntrospectToken(t.Context(), "bad-cid", "bad-csecret", "tok")
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

func TestRevokeToken_Success(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/oauth/revoke" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// verify Basic auth
		authHdr := r.Header.Get("Authorization")
		expected := "Basic " + base64.StdEncoding.EncodeToString([]byte("cid:csecret"))
		if authHdr != expected {
			http.Error(w, "bad auth", http.StatusUnauthorized)
			return
		}
		var body map[string]string
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if body["token"] != "tok-to-revoke" {
			http.Error(w, "wrong token", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"request_id":"req-revoke-1"}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	result, err := client.RevokeToken(t.Context(), "cid", "csecret", "tok-to-revoke")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil || result.RequestID != "req-revoke-1" {
		t.Errorf("RequestID = %v, want %q", result, "req-revoke-1")
	}
}

func TestRevokeToken_Error(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"status":400,"code":"invalid_request","message":"Token not found."}`))
	}))
	defer srv.Close()

	client := api.NewClient("", api.WithBaseURL(srv.URL))
	_, err := client.RevokeToken(t.Context(), "cid", "csecret", "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var apiErr *api.APIError
	if !api.AsAPIError(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T: %v", err, err)
	}
	if apiErr.Code != "invalid_request" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "invalid_request")
	}
}
