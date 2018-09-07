package utils

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

const (
	queryArgDryRun       = "simulate"
	queryArgGenerateOnly = "generate_only"
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
func HasDryRunArg(r *http.Request) bool { return urlQueryHasArg(r.URL, queryArgDryRun) }

// HasGenerateOnlyArg returns whether a URL's query "generate-only" parameter is set to "true".
func HasGenerateOnlyArg(r *http.Request) bool { return urlQueryHasArg(r.URL, queryArgGenerateOnly) }

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

// WriteGenerateStdTxResponse writes response for the generate_only mode.
func WriteGenerateStdTxResponse(w http.ResponseWriter, txBld authctx.TxBuilder, msgs []sdk.Msg) {
	stdMsg, err := txBld.Build(msgs)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := txBld.Codec.MarshalJSON(auth.NewStdTx(stdMsg.Msgs, stdMsg.Fee, nil, stdMsg.Memo))
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Write(output)
	return
}

func urlQueryHasArg(url *url.URL, arg string) bool { return url.Query().Get(arg) == "true" }
