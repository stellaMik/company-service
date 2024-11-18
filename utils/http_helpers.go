package utils

import (
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"net/http"
)

// GetUUIDParam gets and checks a UUID parameter from an HTTP request.
func GetUUIDParam(r *http.Request, param string) (string, error) {
	// Get the parameter from the URL
	uuidParam, ok := mux.Vars(r)[param]
	if !ok {
		return "", errors.New("the parameter " + param + " does not exist")
	}

	// Validate if the parameter is a valid UUID
	_, err := uuid.Parse(uuidParam)
	if err != nil {
		return "", errors.New("the parameter " + param + " is not UUID")
	}

	return uuidParam, nil
}
