package utils

import (
	"encoding/json"
	"net/http"
)

// SendErrorResponse with the appropriate status code, headers, and payload.
func SendErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	SendJSONResponse(w, statusCode, map[string]string{"error": message})
}

// SendJSONResponse with the appropriate status code, headers, and payload.
func SendJSONResponse(w http.ResponseWriter, statusCode int, message interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(message)
}
