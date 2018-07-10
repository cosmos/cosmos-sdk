package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/slashing"
)

func registerTxRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/slashing/unrevoke",
		unrevokeRequestHandlerFn(cdc, kb, ctx),
	).Methods("POST")
}

// Unrevoke TX body
type UnrevokeBody struct {
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	AccountNumber    int64  `json:"account_number"`
	Sequence         int64  `json:"sequence"`
	Gas              int64  `json:"gas"`
	ValidatorAddr    string `json:"validator_addr"`
}

func unrevokeRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m UnrevokeBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		validatorAddr, err := sdk.AccAddressFromBech32(m.ValidatorAddr)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error())))
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), validatorAddr) {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Must use own validator address"))
			return
		}

		ctx = ctx.WithGas(m.Gas)
		ctx = ctx.WithChainID(m.ChainID)
		ctx = ctx.WithAccountNumber(m.AccountNumber)
		ctx = ctx.WithSequence(m.Sequence)

		msg := slashing.NewMsgUnrevoke(validatorAddr)

		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, []sdk.Msg{msg}, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
