package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/pkg/errors"
)

type baseReq struct {
	Name          string `json:"name"`
	Password      string `json:"password"`
	ChainID       string `json:"chain_id"`
	AccountNumber int64  `json:"account_number"`
	Sequence      int64  `json:"sequence"`
	Gas           int64  `json:"gas"`
}

func buildReq(w http.ResponseWriter, r *http.Request, req interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeErr(&w, http.StatusBadRequest, err.Error())
		return err
	}
	err = json.Unmarshal(body, &req)
	if err != nil {
		writeErr(&w, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

func (req baseReq) baseReqValidate(w http.ResponseWriter) bool {
	if len(req.Name) == 0 {
		writeErr(&w, http.StatusUnauthorized, "Name required but not specified")
		return false
	}

	if len(req.Password) == 0 {
		writeErr(&w, http.StatusUnauthorized, "Password required but not specified")
		return false
	}

	if len(req.ChainID) == 0 {
		writeErr(&w, http.StatusUnauthorized, "ChainID required but not specified")
		return false
	}

	if req.AccountNumber < 0 {
		writeErr(&w, http.StatusUnauthorized, "Account Number required but not specified")
		return false
	}

	if req.Sequence < 0 {
		writeErr(&w, http.StatusUnauthorized, "Sequence required but not specified")
		return false
	}
	return true
}

func writeErr(w *http.ResponseWriter, status int, msg string) {
	(*w).WriteHeader(status)
	err := errors.New(msg)
	(*w).Write([]byte(err.Error()))
}

// TODO: Build this function out into a more generic base-request (probably should live in client/lcd)
func signAndBuild(w http.ResponseWriter, ctx context.CoreContext, baseReq baseReq, msg sdk.Msg, cdc *wire.Codec) {
	ctx = ctx.WithAccountNumber(baseReq.AccountNumber)
	ctx = ctx.WithSequence(baseReq.Sequence)
	ctx = ctx.WithChainID(baseReq.ChainID)

	// add gas to context
	ctx = ctx.WithGas(baseReq.Gas)

	txBytes, err := ctx.SignAndBuild(baseReq.Name, baseReq.Password, []sdk.Msg{msg}, cdc)
	if err != nil {
		writeErr(&w, http.StatusUnauthorized, err.Error())
		return
	}

	// send
	res, err := ctx.BroadcastTx(txBytes)
	if err != nil {
		writeErr(&w, http.StatusInternalServerError, err.Error())
		return
	}

	output, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		writeErr(&w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Write(output)
}
