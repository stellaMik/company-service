package utils

import (
	"errors"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

// GetUUIDParam gets and checks a UUID parameter from an HTTP request.
func GetUUIDParam(r *http.Request, param string) (string, error) {
	uuid, ok := mux.Vars(r)[param]
	if !ok {
		return "", errors.New("the parameter " + param + " does not exist")
	}
	if len(strings.Split(uuid, "-")) != 5 {
		return "", errors.New("the parameter " + param + " is not UUID")
	}
	return uuid, nil
}
