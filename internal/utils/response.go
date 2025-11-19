// Package utils has utilities
package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

// SendJSON is an Helper function to send JSON responses
func SendJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// SendError is an Helper function to send error responses
func SendError(w http.ResponseWriter, statusCode int, error string, message string) {
	SendJSON(w, statusCode, ErrorResponse{
		Error:   error,
		Message: message,
	})
}
