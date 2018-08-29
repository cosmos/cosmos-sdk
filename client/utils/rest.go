package utils

import (
	"net/http"
)

// WriteErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func WriteErrorResponse(w *http.ResponseWriter, status int, msg string) {
	(*w).WriteHeader(status)
	(*w).Write([]byte(msg))
}
