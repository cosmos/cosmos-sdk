package rest

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/keyerror"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtxb "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
)

//-----------------------------------------------------------------------------
// Basic HTTP utilities

// ErrorResponse defines the attributes of a JSON error response.
type ErrorResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// NewErrorResponse creates a new ErrorResponse instance.
func NewErrorResponse(code int, msg string) ErrorResponse {
	return ErrorResponse{Code: code, Message: msg}
}

// WriteErrorResponse prepares and writes a HTTP error
// given a status code and an error message.
func WriteErrorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(codec.Cdc.MustMarshalJSON(NewErrorResponse(0, err)))
}

// WriteSimulationResponse prepares and writes an HTTP
// response for transactions simulations.
func WriteSimulationResponse(w http.ResponseWriter, cdc *codec.Codec, gas uint64) {
	gasEst := GasEstimateResponse{GasEstimate: gas}
	resp, err := cdc.MarshalJSON(gasEst)
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
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

//-----------------------------------------------------------------------------
// Building / Sending utilities

// CompleteAndBroadcastTxREST implements a utility function that facilitates
// sending a series of messages in a signed tx. In addition, it will handle
// tx gas simulation and estimation.
//
// NOTE: Also see CompleteAndBroadcastTxCLI.
func CompleteAndBroadcastTxREST(
	w http.ResponseWriter, r *http.Request, cliCtx context.CLIContext,
	baseReq BaseReq, msgs []sdk.Msg, cdc *codec.Codec,
) {

	gasAdj, ok := ParseFloat64OrReturnBadRequest(w, baseReq.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}

	simAndExec, gas, err := client.ParseGas(baseReq.Gas)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// derive the from account address and name from the Keybase
	fromAddress, fromName, err := context.GetFromFields(baseReq.From)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)
	txBldr := authtxb.NewTxBuilder(
		utils.GetTxEncoder(cdc), baseReq.AccountNumber,
		baseReq.Sequence, gas, gasAdj, baseReq.Simulate,
		baseReq.ChainID, baseReq.Memo, baseReq.Fees, baseReq.GasPrices,
	)

	txBldr, err = utils.PrepareTxBuilder(txBldr, cliCtx)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	if baseReq.Simulate || simAndExec {
		if gasAdj < 0 {
			WriteErrorResponse(w, http.StatusBadRequest, client.ErrInvalidGasAdjustment.Error())
			return
		}

		txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if baseReq.Simulate {
			WriteSimulationResponse(w, cdc, txBldr.GetGas())
			return
		}
	}

	txBytes, err := txBldr.BuildAndSign(cliCtx.GetFromName(), baseReq.Password, msgs)
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

// PostProcessResponse performs post processing for a REST response.
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

// WriteGenerateStdTxResponse writes response for the generate only mode.
func WriteGenerateStdTxResponse(
	w http.ResponseWriter, cdc *codec.Codec, cliCtx context.CLIContext, br BaseReq, msgs []sdk.Msg,
) {

	gasAdj, ok := ParseFloat64OrReturnBadRequest(w, br.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}

	simAndExec, gas, err := client.ParseGas(br.Gas)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	txBldr := authtxb.NewTxBuilder(
		utils.GetTxEncoder(cdc), br.AccountNumber, br.Sequence, gas, gasAdj,
		br.Simulate, br.ChainID, br.Memo, br.Fees, br.GasPrices,
	)

	if simAndExec {
		if gasAdj < 0 {
			WriteErrorResponse(w, http.StatusBadRequest, client.ErrInvalidGasAdjustment.Error())
			return
		}

		txBldr, err = utils.EnrichWithGas(txBldr, cliCtx, msgs)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	stdMsg, err := txBldr.Build(msgs)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	output, err := cdc.MarshalJSON(auth.NewStdTx(stdMsg.Msgs, stdMsg.Fee, nil, stdMsg.Memo))
	if err != nil {
		WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
	return
}
