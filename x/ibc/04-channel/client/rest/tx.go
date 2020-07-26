package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	authclient "github.com/KiraCore/cosmos-sdk/x/auth/client"
	"github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/ibc/channels/open-init", channelOpenInitHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc("/ibc/channels/open-try", channelOpenTryHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/open-ack", RestPortID, RestChannelID), channelOpenAckHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/open-confirm", RestPortID, RestChannelID), channelOpenConfirmHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/close-init", RestPortID, RestChannelID), channelCloseInitHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/close-confirm", RestPortID, RestChannelID), channelCloseConfirmHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc("/ibc/packets/receive", recvPacketHandlerFn(clientCtx)).Methods("POST")
}

// channelOpenInitHandlerFn implements a channel open init handler
//
// @Summary Channel open-init
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.ChannelOpenInitReq true "Channel open-init request body"
// @Success 200 {object} PostChannelOpenInit "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/channels/open-init [post]
func channelOpenInitHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChannelOpenInitReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelOpenInit(
			req.PortID,
			req.ChannelID,
			req.Version,
			req.ChannelOrder,
			req.ConnectionHops,
			req.CounterpartyPortID,
			req.CounterpartyChannelID,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// channelOpenTryHandlerFn implements a channel open try handler
//
// @Summary Channel open-try
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.ChannelOpenTryReq true "Channel open-try request body"
// @Success 200 {object} PostChannelOpenTry "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/channels/open-try [post]
func channelOpenTryHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ChannelOpenTryReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelOpenTry(
			req.PortID,
			req.ChannelID,
			req.Version,
			req.ChannelOrder,
			req.ConnectionHops,
			req.CounterpartyPortID,
			req.CounterpartyChannelID,
			req.CounterpartyVersion,
			req.ProofInit,
			req.ProofHeight,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// channelOpenAckHandlerFn implements a channel open ack handler
//
// @Summary Channel open-ack
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Param body body rest.ChannelOpenAckReq true "Channel open-ack request body"
// @Success 200 {object} PostChannelOpenAck "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id}/open-ack [post]
func channelOpenAckHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		var req ChannelOpenAckReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelOpenAck(
			portID,
			channelID,
			req.CounterpartyVersion,
			req.ProofTry,
			req.ProofHeight,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// channelOpenConfirmHandlerFn implements a channel open confirm handler
//
// @Summary Channel open-confirm
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Param body body rest.ChannelOpenConfirmReq true "Channel open-confirm request body"
// @Success 200 {object} PostChannelOpenConfirm "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id}/open-confirm [post]
func channelOpenConfirmHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		var req ChannelOpenConfirmReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelOpenConfirm(
			portID,
			channelID,
			req.ProofAck,
			req.ProofHeight,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// channelCloseInitHandlerFn implements a channel close init handler
//
// @Summary Channel close-init
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Param body body rest.ChannelCloseInitReq true "Channel close-init request body"
// @Success 200 {object} PostChannelCloseInit "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id}/close-init [post]
func channelCloseInitHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		var req ChannelCloseInitReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelCloseInit(
			portID,
			channelID,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// channelCloseConfirmHandlerFn implements a channel close confirm handler
//
// @Summary Channel close-confirm
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Param body body rest.ChannelCloseConfirmReq true "Channel close-confirm request body"
// @Success 200 {object} PostChannelCloseConfirm "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id}/close-confirm [post]
func channelCloseConfirmHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		var req ChannelCloseConfirmReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgChannelCloseConfirm(
			portID,
			channelID,
			req.ProofInit,
			req.ProofHeight,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}

// recvPacketHandlerFn implements a receive packet handler
//
// @Summary Receive packet
// @Tags IBC
// @Accept  json
// @Produce  json
// @Param body body rest.RecvPacketReq true "Receive packet request body"
// @Success 200 {object} PostRecvPacket "OK"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/packets/receive [post]
func recvPacketHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RecvPacketReq
		if !rest.ReadRESTReq(w, r, clientCtx.Codec, &req) {
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
		msg := types.NewMsgRecvPacket(
			req.Packet,
			req.Proofs,
			req.Height,
			fromAddr,
		)

		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
