package cli

import (
	"errors"
	"testing"
)

func TestExitCodeConstants(t *testing.T) {
	t.Parallel()
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitConnection != 1 {
		t.Errorf("ExitConnection = %d, want 1", ExitConnection)
	}
	if ExitAPI != 2 {
		t.Errorf("ExitAPI = %d, want 2", ExitAPI)
	}
	if ExitAuth != 3 {
		t.Errorf("ExitAuth = %d, want 3", ExitAuth)
	}
}

func TestCLIError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *CLIError
		wantMsg  string
		wantCode int
	}{
		{
			name:     "connection error",
			err:      NewCLIError(ExitConnection, "connection refused"),
			wantMsg:  "connection refused",
			wantCode: ExitConnection,
		},
		{
			name:     "api error",
			err:      NewCLIError(ExitAPI, "rate limit exceeded"),
			wantMsg:  "rate limit exceeded",
			wantCode: ExitAPI,
		},
		{
			name:     "auth error",
			err:      NewCLIError(ExitAuth, "invalid token"),
			wantMsg:  "invalid token",
			wantCode: ExitAuth,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.err.Error() != tc.wantMsg {
				t.Errorf("Error() = %q, want %q", tc.err.Error(), tc.wantMsg)
			}
			if tc.err.Code != tc.wantCode {
				t.Errorf("Code = %d, want %d", tc.err.Code, tc.wantCode)
			}
		})
	}
}

func TestExitCodeFromError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		wantCode int
	}{
		{
			name:     "nil error returns success",
			err:      nil,
			wantCode: ExitSuccess,
		},
		{
			name:     "cli error preserves code",
			err:      NewCLIError(ExitAuth, "unauthorized"),
			wantCode: ExitAuth,
		},
		{
			name:     "wrapped cli error preserves code",
			err:      errors.Join(NewCLIError(ExitAPI, "bad request"), errors.New("extra context")),
			wantCode: ExitAPI,
		},
		{
			name:     "generic error returns api exit code",
			err:      errors.New("something went wrong"),
			wantCode: ExitAPI,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := ExitCodeFromError(tc.err)
			if got != tc.wantCode {
				t.Errorf("ExitCodeFromError(%v) = %d, want %d", tc.err, got, tc.wantCode)
			}
		})
	}
}
