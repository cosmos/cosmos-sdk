package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/auth/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/clients", createClientHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/update", RestClientID), updateClientHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/misbehaviour", RestClientID), submitMisbehaviourHandlerFn(cliCtx)).Methods("POST")
}

// createClientHandlerFn implements a create client handler
//
// @Summary Create client
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.CreateClientReq true "Create client request body"
// @Success 200 {object} PostCreateClient "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients [post]
func createClientHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateClientReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		// create the message
		msg := types.NewMsgCreateClient(
			req.ClientID,
			req.ConsensusState.ClientType().String(),
			req.ConsensusState,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// updateClientHandlerFn implements a update client handler
//
// @Summary update client
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param body body rest.UpdateClientReq true "Update client request body"
// @Success 200 {object} PostUpdateClient "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/update [post]
func updateClientHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]

		var req UpdateClientReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		// create the message
		msg := types.NewMsgUpdateClient(
			clientID,
			req.Header,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// submitMisbehaviourHandlerFn implements a submit misbehaviour handler
//
// @Summary Submit misbehaviour
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param body body rest.SubmitMisbehaviourReq true "Submit misbehaviour request body"
// @Success 200 {object} PostSubmitMisbehaviour "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/misbehaviour [post]
func submitMisbehaviourHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]

		var req SubmitMisbehaviourReq
		if !rest.ReadRESTReq(w, r, cliCtx.Codec, &req) {
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

		// create the message
		msg := types.NewMsgSubmitMisbehaviour(
			clientID,
			req.Evidence,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		utils.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
