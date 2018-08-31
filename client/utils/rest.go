package utils

import (
	"fmt"
	"net/http"
)

const (
	queryArgDryRun = "simulate"
)

// WriteErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func WriteErrorResponse(w *http.ResponseWriter, status int, msg string) {
	(*w).WriteHeader(status)
	(*w).Write([]byte(msg))
}

// WriteGasEstimateResponse prepares and writes an HTTP
// response for transactions simulations.
func WriteSimulationResponse(w *http.ResponseWriter, gas int64) {
	(*w).WriteHeader(http.StatusOK)
	(*w).Write([]byte(fmt.Sprintf(`{"gas_estimate":%v}`, gas)))
}

// HasDryRunArg returns true if the request's URL query contains
// the dry run argument and its value is set to "true".
func HasDryRunArg(r *http.Request) bool {
	return r.URL.Query().Get(queryArgDryRun) == "true"
}
