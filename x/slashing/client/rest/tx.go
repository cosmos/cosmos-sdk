package rest

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
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
type UnjailReq struct {
	BaseReq       utils.BaseReq `json:"base_req"`
	ValidatorAddr string        `json:"validator_addr"`
}

// nolint: gocyclo
func unjailRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req UnjailReq
		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			return
		}
		if !req.BaseReq.BaseReqValidate(w) {
			return
		}

		info, err := kb.Get(req.BaseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(req.ValidatorAddr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't decode validator. Error: %s", err.Error()))
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), valAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "Must use own validator address")
			return
		}

		// create the message
		msg := slashing.NewMsgUnjail(valAddr)

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}
