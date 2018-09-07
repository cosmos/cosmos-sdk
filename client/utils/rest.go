package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	auth "github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

const (
	queryArgDryRun       = "simulate"
	queryArgGenerateOnly = "generate_only"
)

//----------------------------------------
// Basic HTTP utilities

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

// ParseInt64OrReturnBadRequest converts s to a float64 value.
func ParseInt64OrReturnBadRequest(w http.ResponseWriter, s string) (n int64, ok bool) {
	var err error
	n, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		err := fmt.Errorf("'%s' is not a valid int64", s)
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return n, false
	}
	return n, true
}

// ParseFloat64OrReturnBadRequest converts s to a float64 value.
// It returns a default value if the string is empty.
func ParseFloat64OrReturnBadRequestDefault(w http.ResponseWriter, s string, defaultIfEmpty float64) (n float64, ok bool) {
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
func WriteGenerateStdTxResponse(w http.ResponseWriter, txBldr authtxb.TxBuilder, msgs []sdk.Msg) {
	stdMsg, err := txBldr.Build(msgs)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}
	output, err := txBldr.Codec.MarshalJSON(auth.NewStdTx(stdMsg.Msgs, stdMsg.Fee, nil, stdMsg.Memo))
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Write(output)
	return
}

func urlQueryHasArg(url *url.URL, arg string) bool { return url.Query().Get(arg) == "true" }

//----------------------------------------
// Building / Sending utilities

type BaseReq struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	ChainID       string `json:"chain_id"`
	AccountNumber int64  `json:"account_number"`
	Sequence      int64  `json:"sequence"`
	Gas           int64  `json:"gas"`
	GasAdjustment string `json:"gas_adjustment"`
}

/*
ReadRESTReq is a simple convenience wrapper that reads the body and
unmarshals to the req interface.

  Usage:
    type SomeReq struct {
      BaseReq            `json:"base_req"`
      CustomField string `json:"custom_field"`
    }
    req := new(SomeReq)
    err := ReadRESTReq(w, r, cdc, req)
*/
func ReadRESTReq(w http.ResponseWriter, r *http.Request, cdc *wire.Codec, req interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return err
	}
	err = cdc.UnmarshalJSON(body, req)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

func (req BaseReq) BaseReqValidate(w http.ResponseWriter) bool {
	if len(req.Name) == 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "Name required but not specified")
		return false
	}

	if len(req.Password) == 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "Password required but not specified")
		return false
	}

	if len(req.ChainID) == 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "ChainID required but not specified")
		return false
	}

	if req.AccountNumber < 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "Account Number required but not specified")
		return false
	}

	if req.Sequence < 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "Sequence required but not specified")
		return false
	}
	return true
}

// CompleteAndBroadcastTxREST implements a utility function that
// facilitates sending a series of messages in a signed
// transaction given a TxBuilder and a QueryContext. It ensures
// that the account exists, has a proper number and sequence
// set. In addition, it builds and signs a transaction with the
// supplied messages.  Finally, it broadcasts the signed
// transaction to a node.
// NOTE: Also see CompleteAndBroadcastTxCli.
// NOTE: Also see x/stake/client/rest/tx.go delegationsRequestHandlerFn.
func CompleteAndBroadcastTxREST(w http.ResponseWriter, r *http.Request, cliCtx context.CLIContext, baseReq BaseReq, msgs []sdk.Msg, cdc *wire.Codec) {
	var err error
	txBldr := authtxb.TxBuilder{
		Codec:         cdc,
		AccountNumber: baseReq.AccountNumber,
		Sequence:      baseReq.Sequence,
		ChainID:       baseReq.ChainID,
		Gas:           baseReq.Gas,
	}

	adjustment, ok := ParseFloat64OrReturnBadRequestDefault(w, baseReq.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}
	cliCtx = cliCtx.WithGasAdjustment(adjustment)

	if HasDryRunArg(r) || baseReq.Gas == 0 {
		newCtx, err := EnrichCtxWithGas(txBldr, cliCtx, baseReq.Name, msgs)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if HasDryRunArg(r) {
			WriteSimulationResponse(w, txBldr.Gas)
			return
		}
		txBldr = newCtx
	}

	if HasGenerateOnlyArg(r) {
		WriteGenerateStdTxResponse(w, txBldr, msgs)
		return
	}

	txBytes, err := txBldr.BuildAndSign(baseReq.Name, baseReq.Password, msgs)
	if err != nil {
		WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(output)
}
