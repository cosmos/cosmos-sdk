package rest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"
	"github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc(
		"/slashing/unjail",
		unjailRequestHandlerFn(cdc, kb, cliCtx),
	).Methods("POST")
}

// Unjail TX body
type UnjailBody struct {
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	AccountNumber    int64  `json:"account_number"`
	Sequence         int64  `json:"sequence"`
	Gas              int64  `json:"gas"`
	ValidatorAddr    string `json:"validator_addr"`
}

func unjailRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var m UnjailBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusBadRequest, err.Error())
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusBadRequest, err.Error())
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusUnauthorized, err.Error())
			return
		}

		validatorAddr, err := sdk.AccAddressFromBech32(m.ValidatorAddr)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), validatorAddr) {
			utils.WriteErrorResponse(&w, http.StatusUnauthorized, "Must use own validator address")
			return
		}

		txCtx := authctx.TxContext{
			Codec:         cdc,
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
			Gas:           m.Gas,
		}

		msg := slashing.NewMsgUnjail(validatorAddr)

		if m.Gas == 0 {
			newCtx, err := utils.EnrichCtxWithGas(txCtx, cliCtx, m.LocalAccountName, m.Password, []sdk.Msg{msg})
			if err != nil {
				utils.WriteErrorResponse(&w, http.StatusInternalServerError, err.Error())
				return
			}
			txCtx = newCtx
		}

		txBytes, err := txCtx.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusUnauthorized, "Must use own validator address")
			return
		}

		res, err := cliCtx.BroadcastTx(txBytes)
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusInternalServerError, err.Error())
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			utils.WriteErrorResponse(&w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Write(output)
	}
}
