package utils

import (
	"fmt"
	"net/http"
	"strconv"
)

const (
	queryArgDryRun = "simulate"
)

// WriteErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func WriteErrorResponse(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
}

// WriteGasEstimateResponse prepares and writes an HTTP
// response for transactions simulations.
func WriteSimulationResponse(w http.ResponseWriter, gas int64) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"gas_estimate":%v}`, gas)))
}

// HasDryRunArg returns true if the request's URL query contains
// the dry run argument and its value is set to "true".
func HasDryRunArg(r *http.Request) bool {
	return r.URL.Query().Get(queryArgDryRun) == "true"
}

// ParseFloat64OrReturnBadRequest converts s to a float64 value. It returns a default
// value if the string is empty. Write
func ParseFloat64OrReturnBadRequest(w http.ResponseWriter, s string, defaultIfEmpty float64) (n float64, ok bool) {
	if len(s) == 0 {
		return defaultIfEmpty, true
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return n, false
	}
	return n, true
}
