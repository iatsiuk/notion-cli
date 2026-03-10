package cli

import "errors"

const (
	ExitSuccess    = 0
	ExitConnection = 1
	ExitAPI        = 2
	ExitAuth       = 3
)

// CLIError is an error with an associated exit code.
type CLIError struct {
	Code    int
	message string
}

func (e *CLIError) Error() string {
	return e.message
}

// NewCLIError creates a CLIError with the given exit code and message.
func NewCLIError(code int, message string) *CLIError {
	return &CLIError{Code: code, message: message}
}

// ExitCodeFromError returns the exit code for the given error.
// Returns ExitSuccess for nil, ExitAPI for unknown errors.
func ExitCodeFromError(err error) int {
	if err == nil {
		return ExitSuccess
	}
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		return cliErr.Code
	}
	return ExitAPI
}
