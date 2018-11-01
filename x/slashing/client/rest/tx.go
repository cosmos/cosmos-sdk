package rest

import (
	"bytes"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"

	"github.com/gorilla/mux"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec,
	kbInit func() (keys.Keybase, error)) {
	r.HandleFunc(
		"/slashing/validators/{validatorAddr}/unjail",
		unjailRequestHandlerFn(cdc, kbInit, cliCtx),
	).Methods("POST")
}

// Unjail TX body
type UnjailReq struct {
	BaseReq utils.BaseReq `json:"base_req"`
}

func unjailRequestHandlerFn(cdc *codec.Codec, kbInit func() (keys.Keybase, error),
	cliCtx context.CLIContext) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		bech32validator := vars["validatorAddr"]

		var req UnjailReq
		err := utils.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			return
		}

		baseReq := req.BaseReq.Sanitize()
		if !baseReq.ValidateBasic(w) {
			return
		}

		kb, err := kbInit()
		if err != nil {
			panic(err)
		}
		info, err := kb.Get(baseReq.Name)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, err.Error())
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		if !bytes.Equal(info.GetPubKey().Address(), valAddr) {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "must use own validator address")
			return
		}

		msg := slashing.NewMsgUnjail(valAddr)
		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, baseReq, []sdk.Msg{msg}, cdc)
	}
}
