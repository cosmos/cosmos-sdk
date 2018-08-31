package rest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
)

type baseReq struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	ChainID       string `json:"chain_id"`
	AccountNumber int64  `json:"account_number"`
	Sequence      int64  `json:"sequence"`
	Gas           int64  `json:"gas"`
	GasAdjustment string `json:"gas_adjustment"`
}

func buildReq(w http.ResponseWriter, r *http.Request, cdc *wire.Codec, req interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return err
	}
	err = cdc.UnmarshalJSON(body, req)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

func (req baseReq) baseReqValidate(w http.ResponseWriter) bool {
	if len(req.Name) == 0 {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Name required but not specified")
		return false
	}

	if len(req.Password) == 0 {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Password required but not specified")
		return false
	}

	if len(req.ChainID) == 0 {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "ChainID required but not specified")
		return false
	}

	if req.AccountNumber < 0 {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Account Number required but not specified")
		return false
	}

	if req.Sequence < 0 {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "Sequence required but not specified")
		return false
	}
	return true
}

// TODO: Build this function out into a more generic base-request
// (probably should live in client/lcd).
func signAndBuild(w http.ResponseWriter, r *http.Request, cliCtx context.CLIContext, baseReq baseReq, msg sdk.Msg, cdc *wire.Codec) {
	var err error
	txCtx := authctx.TxContext{
		Codec:         cdc,
		AccountNumber: baseReq.AccountNumber,
		Sequence:      baseReq.Sequence,
		ChainID:       baseReq.ChainID,
		Gas:           baseReq.Gas,
	}

	adjustment, ok := utils.ParseFloat64OrReturnBadRequest(w, baseReq.GasAdjustment, client.DefaultGasAdjustment)
	if !ok {
		return
	}
	cliCtx = cliCtx.WithGasAdjustment(adjustment)

	if utils.HasDryRunArg(r) || baseReq.Gas == 0 {
		newCtx, err := utils.EnrichCtxWithGas(txCtx, cliCtx, baseReq.Name, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}
		if utils.HasDryRunArg(r) {
			utils.WriteSimulationResponse(w, txCtx.Gas)
			return
		}
		txCtx = newCtx
	}
	txBytes, err := txCtx.BuildAndSign(baseReq.Name, baseReq.Password, []sdk.Msg{msg})
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	res, err := cliCtx.BroadcastTx(txBytes)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	output, err := wire.MarshalJSONIndent(cdc, res)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(output)
}

func parseInt64OrReturnBadRequest(s string, w http.ResponseWriter) (n int64, ok bool) {
	var err error
	n, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		err := fmt.Errorf("'%s' is not a valid int64", s)
		w.Write([]byte(err.Error()))
		return 0, false
	}
	return n, true
}
