package api

import (
	"encoding/json"
	"fmt"
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

// parseAPIError parses a Notion API error response body.
func parseAPIError(status int, body []byte) error {
	var apiErr APIError
	if err := json.Unmarshal(body, &apiErr); err != nil {
		return fmt.Errorf("notion api error %d: %s", status, string(body))
	}
	apiErr.Status = status
	return &apiErr
}
