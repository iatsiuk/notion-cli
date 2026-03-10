package api

import (
	"encoding/json"
	"errors"
	"fmt"
)

// exit code values matching cli.Exit* constants (avoid import cycle).
const (
	exitAPI        = 2
	exitAuth       = 3
	exitConnection = 1
)

// APIError represents an error response from the Notion API.
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("notion api error %d (%s): %s", e.Status, e.Code, e.Message)
}

// ExitCode returns the CLI exit code for this error.
// 401 and 403 map to ExitAuth; all other API errors map to ExitAPI.
func (e *APIError) ExitCode() int {
	if e.Status == 401 || e.Status == 403 {
		return exitAuth
	}
	return exitAPI
}

// ConnectionError wraps a network-level error (e.g. connection refused).
type ConnectionError struct {
	cause error
}

func (e *ConnectionError) Error() string {
	return fmt.Sprintf("connection error: %s", e.cause)
}

func (e *ConnectionError) Unwrap() error { return e.cause }

// ExitCode returns ExitConnection for network-level errors.
func (e *ConnectionError) ExitCode() int { return exitConnection }

// AsAPIError reports whether any error in err's chain is *APIError
// and sets target to that value.
func AsAPIError(err error, target **APIError) bool {
	return errors.As(err, target)
}

// AsConnectionError reports whether any error in err's chain is *ConnectionError.
func AsConnectionError(err error, target **ConnectionError) bool {
	return errors.As(err, target)
}

// parseAPIError parses a Notion API error response body.
// Always returns *APIError, even when the body is not valid JSON.
func parseAPIError(status int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return &APIError{Status: status, Message: string(body)}
	}
	apiErr.Status = status
	return &apiErr
}
