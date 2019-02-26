package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"

	"github.com/cosmos/cosmos-sdk/x/bank"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc("/bank/accounts/{address}/transfers", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

// SendReq defines the properties of a send request's body.
type SendReq struct {
	BaseReq rest.BaseReq `json:"base_req"`
	Amount  sdk.Coins    `json:"amount"`
}

var msgCdc = codec.New()

func init() {
	bank.RegisterCodec(msgCdc)
}

// SendRequestHandlerFn - http request handler to send coins to a address.
func SendRequestHandlerFn(cdc *codec.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32Addr := vars["address"]

		toAddr, err := sdk.AccAddressFromBech32(bech32Addr)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		var req SendReq
		if !rest.ReadRESTReq(w, r, cdc, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		msg := bank.NewMsgSend(fromAddr, toAddr, req.Amount)
		clientrest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
