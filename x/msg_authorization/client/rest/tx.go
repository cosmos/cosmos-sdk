package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/internal/types"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/msg_authorization/grant", grantHandler(cliCtx)).Methods("POST")
	r.HandleFunc("/msg_authorization/revoke", revokeHandler(cliCtx)).Methods("POST")
}

type GrantRequest struct {
	BaseReq    rest.BaseReq     `json:"base_req" yaml:"base_req"`
	Granter    sdk.AccAddress   `json:"granter"`
	Grantee    sdk.AccAddress   `json:"grantee"`
	Capability types.Capability `json:"capability"`
	Expiration time.Time        `json:"expiration"`
}

type RevokeRequest struct {
	BaseReq           rest.BaseReq   `json:"base_req" yaml:"base_req"`
	Granter           sdk.AccAddress `json:"granter"`
	Grantee           sdk.AccAddress `json:"grantee"`
	CapabilityMsgType sdk.Msg        `json:"capability_msg_type"`
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

		msg := types.NewMsgGrantAuthorization(req.Granter, req.Grantee, req.Capability, req.Expiration)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
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

		msg := types.NewMsgRevokeAuthorization(req.Granter, req.Grantee, req.CapabilityMsgType)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
