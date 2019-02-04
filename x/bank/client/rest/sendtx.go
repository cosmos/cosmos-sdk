package rest

import (
	"net/http"

	"github.com/cosmos/cosmos-sdk/client/rest"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankclient "github.com/cosmos/cosmos-sdk/x/bank/client"

	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec, kb keys.Keybase) {
	r.HandleFunc("/bank/accounts/{address}/transfers", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
	r.HandleFunc("/tx/broadcast", BroadcastTxRequestHandlerFn(cdc, cliCtx)).Methods("POST")
}

type sendReq struct {
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

		var req sendReq
		err = rest.ReadRESTReq(w, r, cdc, &req)
		if err != nil {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		if req.BaseReq.GenerateOnly {
			// When generate only is supplied, the from field must be a valid Bech32
			// address.
			fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
			if err != nil {
				rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
				return
			}

			msg := bankclient.CreateMsg(fromAddr, toAddr, req.Amount)
			rest.WriteGenerateStdTxResponse(w, cdc, cliCtx, req.BaseReq, []sdk.Msg{msg})
			return
		}

		// derive the from account address and name from the Keybase
		fromAddress, fromName, err := context.GetFromFields(req.BaseReq.From)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		cliCtx = cliCtx.WithFromName(fromName).WithFromAddress(fromAddress)
		msg := bankclient.CreateMsg(cliCtx.GetFromAddress(), toAddr, req.Amount)

		rest.CompleteAndBroadcastTxREST(w, r, cliCtx, req.BaseReq, []sdk.Msg{msg}, cdc)
	}
}
