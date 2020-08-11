package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/ibc/clients/tendermint", createClientHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/update", RestClientID), updateClientHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/misbehaviour", RestClientID), submitMisbehaviourHandlerFn(clientCtx)).Methods("POST")
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
// @Router /ibc/clients/tendermint [post]
func createClientHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateClientReq
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
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
		msg := ibctmtypes.NewMsgCreateClient(
			req.ClientID, req.Header, req.TrustLevel,
			req.TrustingPeriod, req.UnbondingPeriod, req.MaxClockDrift,
			req.ProofSpecs, fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
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
func updateClientHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]

		var req UpdateClientReq
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
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
		msg := ibctmtypes.NewMsgUpdateClient(
			clientID,
			req.Header,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}

// submitMisbehaviourHandlerFn implements a submit misbehaviour handler
//
// @Summary Submit misbehaviour
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.SubmitMisbehaviourReq true "Submit misbehaviour request body"
// @Success 200 {object} PostSubmitMisbehaviour "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/misbehaviour [post]
func submitMisbehaviourHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SubmitMisbehaviourReq
		if !rest.ReadRESTReq(w, r, clientCtx.LegacyAmino, &req) {
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
		msg := ibctmtypes.NewMsgSubmitClientMisbehaviour(req.Evidence, fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		tx.WriteGeneratedTxResponse(clientCtx, w, req.BaseReq, msg)
	}
}
