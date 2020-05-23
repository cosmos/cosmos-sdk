package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/connections/open-init", connectionOpenInitHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/ibc/connections/open-try", connectionOpenTryHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}/open-ack", RestConnectionID), connectionOpenAckHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}/open-confirm", RestConnectionID), connectionOpenConfirmHandlerFn(cliCtx)).Methods("POST")
}

// connectionOpenInitHandlerFn implements a connection open init handler
//
// @Summary Connection open-init
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.ConnectionOpenInitReq true "Connection open-init request body"
// @Success 200 {object} PostConnectionOpenInit "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections/open-init [post]
func connectionOpenInitHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ConnectionOpenInitReq
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
		msg, err := types.NewMsgConnectionOpenInit(
			req.ConnectionID, req.ClientID, req.CounterpartyConnectionID,
			req.CounterpartyClientID, req.CounterpartyPrefix, fromAddr,
		)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// connectionOpenTryHandlerFn implements a connection open try handler
//
// @Summary Connection open-try
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.ConnectionOpenTryReq true "Connection open-try request body"
// @Success 200 {object} PostConnectionOpenTry "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections/open-try [post]
func connectionOpenTryHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ConnectionOpenTryReq
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
		msg, err := types.NewMsgConnectionOpenTry(
			req.ConnectionID, req.ClientID, req.CounterpartyConnectionID,
			req.CounterpartyClientID, req.CounterpartyPrefix, req.CounterpartyVersions,
			req.ProofInit, req.ProofConsensus, req.ProofHeight,
			req.ConsensusHeight, fromAddr,
		)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// connectionOpenAckHandlerFn implements a connection open ack handler
//
// @Summary Connection open-ack
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param connection-id path string true "Connection ID"
// @Param body body rest.ConnectionOpenAckReq true "Connection open-ack request body"
// @Success 200 {object} PostConnectionOpenAck "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid connection id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections/{connection-id}/open-ack [post]
func connectionOpenAckHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]

		var req ConnectionOpenAckReq
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
		msg, err := types.NewMsgConnectionOpenAck(
			connectionID, req.ProofTry, req.ProofConsensus, req.ProofHeight,
			req.ConsensusHeight, req.Version, fromAddr,
		)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// connectionOpenConfirmHandlerFn implements a connection open confirm handler
//
// @Summary Connection open-confirm
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param connection-id path string true "Connection ID"
// @Param body body rest.ConnectionOpenConfirmReq true "Connection open-confirm request body"
// @Success 200 {object} PostConnectionOpenConfirm "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid connection id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections/{connection-id}/open-confirm [post]
func connectionOpenConfirmHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]

		var req ConnectionOpenConfirmReq
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
		msg, err := types.NewMsgConnectionOpenConfirm(
			connectionID, req.ProofAck, req.ProofHeight, fromAddr,
		)

		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		}

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
