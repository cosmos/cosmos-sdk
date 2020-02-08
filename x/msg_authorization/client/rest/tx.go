package rest

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/msg_authorization/grant", grantHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/msg_authorization/revoke", revokeHandler(cliCtx)).Methods("POST")
}

type GrantRequest struct {
	BaseReq       rest.BaseReq        `json:"base_req" yaml:"base_req"`
	Granter       sdk.AccAddress      `json:"granter"`
	Grantee       sdk.AccAddress      `json:"grantee"`
	Authorization types.Authorization `json:"authorization"`
	Expiration    time.Time           `json:"expiration"`
}

type RevokeRequest struct {
	BaseReq              rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Granter              sdk.AccAddress `json:"granter"`
	Grantee              sdk.AccAddress `json:"grantee"`
	AuthorizationMsgType sdk.Msg        `json:"authorization_msg_type"`
}

func grantHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GrantRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgGrantAuthorization(req.Granter, req.Grantee, req.Authorization, req.Expiration)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

func revokeHandler(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RevokeRequest

		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
			return
		}

		req.BaseReq = req.BaseReq.Sanitize()
		if !req.BaseReq.ValidateBasic(w) {
			return
		}

		msg := types.NewMsgRevokeAuthorization(req.Granter, req.Grantee, req.AuthorizationMsgType)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
