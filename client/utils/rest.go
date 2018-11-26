package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
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
func WriteErrorResponse(w http.ResponseWriter, status int, err string) {
	w.WriteHeader(status)
	w.Write([]byte(err))
}

// WriteSimulationResponse prepares and writes an HTTP
// response for transactions simulations.
func WriteSimulationResponse(w http.ResponseWriter, gas uint64) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"gas_estimate":%v}`, gas)))
}

// HasDryRunArg returns true if the request's URL query contains the dry run
// argument and its value is set to "true".
func HasDryRunArg(r *http.Request) bool {
	return urlQueryHasArg(r.URL, queryArgDryRun)
}

// HasGenerateOnlyArg returns whether a URL's query "generate-only" parameter
// is set to "true".
func HasGenerateOnlyArg(r *http.Request) bool {
	return urlQueryHasArg(r.URL, queryArgGenerateOnly)
}

// ParseInt64OrReturnBadRequest converts s to a int64 value.
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

// ParseUint64OrReturnBadRequest converts s to a uint64 value.
func ParseUint64OrReturnBadRequest(w http.ResponseWriter, s string) (n uint64, ok bool) {
	var err error

	n, err = strconv.ParseUint(s, 10, 64)
	if err != nil {
		err := fmt.Errorf("'%s' is not a valid uint64", s)
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return n, false
	}

	return n, true
}

// ParseFloat64OrReturnBadRequest converts s to a float64 value. It returns a
// default value, defaultIfEmpty, if the string is empty.
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

// BaseReq defines a structure that can be embedded in other request structures
// that all share common "base" fields.
type BaseReq struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	ChainID       string `json:"chain_id"`
	AccountNumber uint64 `json:"account_number"`
	Sequence      uint64 `json:"sequence"`
	Gas           string `json:"gas"`
	GasAdjustment string `json:"gas_adjustment"`
}

// Sanitize performs basic sanitization on a BaseReq object.
func (br BaseReq) Sanitize() BaseReq {
	return BaseReq{
		Name:          strings.TrimSpace(br.Name),
		Password:      strings.TrimSpace(br.Password),
		ChainID:       strings.TrimSpace(br.ChainID),
		Gas:           strings.TrimSpace(br.Gas),
		GasAdjustment: strings.TrimSpace(br.GasAdjustment),
		AccountNumber: br.AccountNumber,
		Sequence:      br.Sequence,
	}
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
func ReadRESTReq(w http.ResponseWriter, r *http.Request, cdc *codec.Codec, req interface{}) error {
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

// ValidateBasic performs basic validation of a BaseReq. If custom validation
// logic is needed, the implementing request handler should perform those
// checks manually.
func (br BaseReq) ValidateBasic(w http.ResponseWriter) bool {
	switch {
	case len(br.Name) == 0:
		WriteErrorResponse(w, http.StatusUnauthorized, "name required but not specified")
		return false

	case len(br.Password) == 0:
		WriteErrorResponse(w, http.StatusUnauthorized, "password required but not specified")
		return false

	case len(br.ChainID) == 0:
		WriteErrorResponse(w, http.StatusUnauthorized, "chainID required but not specified")
		return false
	}

	return true
}

// CompleteAndBroadcastTxREST implements a utility function that facilitates
// sending a series of messages in a signed transaction given a TxBuilder and a
// QueryContext. It ensures that the account exists, has a proper number and
// sequence set. In addition, it builds and signs a transaction with the
// supplied messages. Finally, it broadcasts the signed transaction to a node.
//
// NOTE: Also see CompleteAndBroadcastTxCli.
// NOTE: Also see x/stake/client/rest/tx.go delegationsRequestHandlerFn.
func CompleteAndBroadcastTxREST(w http.ResponseWriter, r *http.Request, cliCtx context.CLIContext, baseReq BaseReq, msgs []sdk.Msg, cdc *codec.Codec) {
	simulateGas, gas, err := client.ReadGasFlag(baseReq.Gas)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	adjustment, ok := ParseFloat64OrReturnBadRequest(w, baseReq.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}

	txBldr := authtxb.TxBuilder{
		Codec:         cdc,
		Gas:           gas,
		GasAdjustment: adjustment,
		SimulateGas:   simulateGas,
		ChainID:       baseReq.ChainID,
		AccountNumber: baseReq.AccountNumber,
		Sequence:      baseReq.Sequence,
	}

	if HasDryRunArg(r) || txBldr.SimulateGas {
		newBldr, err := EnrichCtxWithGas(txBldr, cliCtx, baseReq.Name, msgs)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if HasDryRunArg(r) {
			WriteSimulationResponse(w, newBldr.Gas)
			return
		}

		txBldr = newBldr
	}

	if HasGenerateOnlyArg(r) {
		WriteGenerateStdTxResponse(w, txBldr, msgs)
		return
	}

	txBytes, err := txBldr.BuildAndSign(baseReq.Name, baseReq.Password, msgs)
	if keyerror.IsErrKeyNotFound(err) {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	} else if keyerror.IsErrWrongPassword(err) {
		WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	} else if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	PostProcessResponse(w, cdc, res, cliCtx.Indent)
}

// PostProcessResponse performs post process for rest response
func PostProcessResponse(w http.ResponseWriter, cdc *codec.Codec, response interface{}, indent bool) {
	var output []byte
	switch response.(type) {
	default:
		var err error
		if indent {
			output, err = cdc.MarshalJSONIndent(response, "", "  ")
		} else {
			output, err = cdc.MarshalJSON(response)
		}
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	case []byte:
		output = response.([]byte)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}
