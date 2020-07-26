package rest

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/tx"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	authclient "github.com/KiraCore/cosmos-sdk/x/auth/client"
	"github.com/KiraCore/cosmos-sdk/x/slashing/types"
)

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/slashing/validators/{validatorAddr}/unjail", NewUnjailRequestHandlerFn(clientCtx)).Methods("POST")
}

// Unjail TX body
type UnjailReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

// NewUnjailRequestHandlerFn returns an HTTP REST handler for creating a MsgUnjail
// transaction.
func NewUnjailRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		bech32Validator := vars["validatorAddr"]

		var req UnjailReq
		if !rest.ReadRESTReq(w, r, clientCtx.JSONMarshaler, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(bech32Validator)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, valAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own validator address")
			return
		}

		msg := types.NewMsgUnjail(valAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}
		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------
// ref: https://github.com/KiraCore/cosmos-sdk/issues/5864
func registerTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(
		"/slashing/validators/{validatorAddr}/unjail",
		unjailRequestHandlerFn(clientCtx),
	).Methods("POST")
}

func unjailRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		bech32validator := vars["validatorAddr"]

		var req UnjailReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		valAddr, err := sdk.ValAddressFromBech32(bech32validator)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		fromAddr, err := sdk.AccAddressFromBech32(req.BaseReq.From)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		if !bytes.Equal(fromAddr, valAddr) {
			rest.WriteErrorResponse(w, http.StatusUnauthorized, "must use own validator address")
			return
		}

		msg := types.NewMsgUnjail(valAddr)
		if rest.CheckBadRequestError(w, msg.ValidateBasic()) {
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
