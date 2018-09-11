package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/ibc"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/ibc/{destchain}/{address}/send", TransferRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

type transferReq struct {
	BaseReq utils.BaseReq `json:"base_req"`
	// Fees             sdk.Coin  `json="fees"`
	Amount sdk.Coins `json:"amount"`
}

// TransferRequestHandler - http request handler to transfer coins to a address
// on a different chain via IBC
// nolint: gocyclo
func TransferRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		destChainID := vars["destchain"]
		bech32addr := vars["address"]

		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var req transferReq
		err = utils.ReadRESTReq(w, r, cdc, &req)
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

		// create the message
		packet := ibc.NewIBCPacket(sdk.AccAddress(info.GetPubKey().Address()), to, req.Amount, req.BaseReq.ChainID, destChainID)
		msg := ibc.IBCTransferMsg{packet}

		utils.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}
