package api_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"notion-cli/internal/api"
	"notion-cli/internal/cli"
)

func TestAPIErrorParsing(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		body     string
		wantCode string
		wantMsg  string
	}{
		{
			name:     "400 bad request",
			status:   http.StatusBadRequest,
			body:     `{"object":"error","status":400,"code":"validation_error","message":"invalid input"}`,
			wantCode: "validation_error",
			wantMsg:  "invalid input",
		},
		{
			name:     "401 unauthorized",
			status:   http.StatusUnauthorized,
			body:     `{"object":"error","status":401,"code":"unauthorized","message":"API token is invalid"}`,
			wantCode: "unauthorized",
			wantMsg:  "API token is invalid",
		},
		{
			name:     "404 not found",
			status:   http.StatusNotFound,
			body:     `{"object":"error","status":404,"code":"object_not_found","message":"Could not find page"}`,
			wantCode: "object_not_found",
			wantMsg:  "Could not find page",
		},
		{
			name:     "429 rate limited",
			status:   http.StatusTooManyRequests,
			body:     `{"object":"error","status":429,"code":"rate_limited","message":"rate limit exceeded"}`,
			wantCode: "rate_limited",
			wantMsg:  "rate limit exceeded",
		},
		{
			name:     "500 server error",
			status:   http.StatusInternalServerError,
			body:     `{"object":"error","status":500,"code":"internal_server_error","message":"internal error"}`,
			wantCode: "internal_server_error",
			wantMsg:  "internal error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			client := api.NewClient("token", api.WithBaseURL(srv.URL))
			_, err := client.Get(t.Context(), "/v1/pages/abc", nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			var apiErr *api.APIError
			if !api.AsAPIError(err, &apiErr) {
				t.Fatalf("expected *api.APIError, got %T: %v", err, err)
			}
			if apiErr.Status != tc.status {
				t.Errorf("Status = %d, want %d", apiErr.Status, tc.status)
			}
			if apiErr.Code != tc.wantCode {
				t.Errorf("Code = %q, want %q", apiErr.Code, tc.wantCode)
			}
			if apiErr.Message != tc.wantMsg {
				t.Errorf("Message = %q, want %q", apiErr.Message, tc.wantMsg)
			}
		})
	}
}

func TestAPIErrorExitCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   int
		wantExit int
	}{
		{"400 -> ExitAPI", http.StatusBadRequest, cli.ExitAPI},
		{"401 -> ExitAuth", http.StatusUnauthorized, cli.ExitAuth},
		{"404 -> ExitAPI", http.StatusNotFound, cli.ExitAPI},
		{"429 -> ExitAPI", http.StatusTooManyRequests, cli.ExitAPI},
		{"500 -> ExitAPI", http.StatusInternalServerError, cli.ExitAPI},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				_, _ = fmt.Fprintf(w, `{"object":"error","status":%d,"code":"test","message":"test"}`, tc.status)
			}))
			defer srv.Close()

			client := api.NewClient("token", api.WithBaseURL(srv.URL))
			_, err := client.Get(t.Context(), "/v1/pages/abc", nil)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			got := cli.ExitCodeFromError(err)
			if got != tc.wantExit {
				t.Errorf("ExitCodeFromError() = %d, want %d (err: %v)", got, tc.wantExit, err)
			}
		})
	}
}

func TestConnectionError(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	srvURL := srv.URL
	srv.Close() // close immediately to cause connection refused

	client := api.NewClient("token", api.WithBaseURL(srvURL))
	_, err := client.Get(context.Background(), "/v1/pages/abc", nil)
	if err == nil {
		t.Fatal("expected connection error, got nil")
	}

	got := cli.ExitCodeFromError(err)
	if got != cli.ExitConnection {
		t.Errorf("ExitCodeFromError() = %d, want %d (ExitConnection)", got, cli.ExitConnection)
	}
}
