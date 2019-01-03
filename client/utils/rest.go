package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
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
func WriteGenerateStdTxResponse(w http.ResponseWriter, cdc *codec.Codec, txBldr authtxb.TxBuilder, msgs []sdk.Msg) {
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

	w.Write(output)
	return
}

//----------------------------------------
// Building / Sending utilities

// BaseReq defines a structure that can be embedded in other request structures
// that all share common "base" fields.
type BaseReq struct {
	Name          string    `json:"name"`
	Password      string    `json:"password"`
	Memo          string    `json:"memo"`
	ChainID       string    `json:"chain_id"`
	AccountNumber uint64    `json:"account_number"`
	Sequence      uint64    `json:"sequence"`
	Fees          sdk.Coins `json:"fees"`
	Gas           string    `json:"gas"`
	GasAdjustment string    `json:"gas_adjustment"`
	GenerateOnly  bool      `json:"generate_only"`
	Simulate      bool      `json:"simulate"`
}

// NewBaseReq creates a new basic request instance and sanitizes its values
func NewBaseReq(
	name, password, memo, chainID string, gas, gasAdjustment string,
	accNumber, seq uint64, fees sdk.Coins, genOnly, simulate bool) BaseReq {

	return BaseReq{
		Name:          strings.TrimSpace(name),
		Password:      password,
		Memo:          strings.TrimSpace(memo),
		ChainID:       strings.TrimSpace(chainID),
		Fees:          fees,
		Gas:           strings.TrimSpace(gas),
		GasAdjustment: strings.TrimSpace(gasAdjustment),
		AccountNumber: accNumber,
		Sequence:      seq,
		GenerateOnly:  genOnly,
		Simulate:      simulate,
	}
}

// Sanitize performs basic sanitization on a BaseReq object.
func (br BaseReq) Sanitize() BaseReq {
	newBr := NewBaseReq(
		br.Name, br.Password, br.Memo, br.ChainID, br.Gas, br.GasAdjustment,
		br.AccountNumber, br.Sequence, br.Fees, br.GenerateOnly, br.Simulate,
	)
	return newBr
}

// ValidateBasic performs basic validation of a BaseReq. If custom validation
// logic is needed, the implementing request handler should perform those
// checks manually.
func (br BaseReq) ValidateBasic(w http.ResponseWriter) bool {
	if !br.GenerateOnly && !br.Simulate {
		switch {
		case len(br.Password) == 0:
			WriteErrorResponse(w, http.StatusUnauthorized, "password required but not specified")
			return false
		case len(br.ChainID) == 0:
			WriteErrorResponse(w, http.StatusUnauthorized, "chain-id required but not specified")
			return false
		case !br.Fees.IsValid():
			WriteErrorResponse(w, http.StatusPaymentRequired, "invalid or insufficient fees")
			return false
		}
	}
	if len(br.Name) == 0 {
		WriteErrorResponse(w, http.StatusUnauthorized, "name required but not specified")
		return false
	}
	return true
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
		WriteErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("failed to decode JSON payload: %s", err))
		return err
	}

	return nil
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
	gasAdjustment, ok := ParseFloat64OrReturnBadRequest(w, baseReq.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}

	simulateAndExecute, gas, err := client.ParseGas(baseReq.Gas)
	if err != nil {
		WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	txBldr := authtxb.NewTxBuilder(GetTxEncoder(cdc), baseReq.AccountNumber,
		baseReq.Sequence, gas, gasAdjustment, baseReq.Simulate,
		baseReq.ChainID, baseReq.Memo, baseReq.Fees)

	if baseReq.Simulate || simulateAndExecute {
		if gasAdjustment < 0 {
			WriteErrorResponse(w, http.StatusBadRequest, "gas adjustment must be a positive float")
			return
		}

		txBldr, err = EnrichCtxWithGas(txBldr, cliCtx, baseReq.Name, msgs)
		if err != nil {
			WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if baseReq.Simulate {
			WriteSimulationResponse(w, txBldr.GetGas())
			return
		}
	}

	if baseReq.GenerateOnly {
		WriteGenerateStdTxResponse(w, cdc, txBldr, msgs)
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
