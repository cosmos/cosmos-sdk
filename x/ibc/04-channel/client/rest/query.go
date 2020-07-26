package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/client/flags"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	"github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/client/utils"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}", RestPortID, RestChannelID), queryChannelHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/client_state", RestPortID, RestChannelID), queryChannelClientStateHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/next_sequence_receive", RestPortID, RestChannelID), queryNextSequenceRecvHandlerFn(clientCtx)).Methods("GET")
}

func queryChannelHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		channelRes, err := utils.QueryChannel(clientCtx, portID, channelID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(channelRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, channelRes)
	}
}

func queryChannelClientStateHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		clientState, height, err := utils.QueryChannelClientState(clientCtx, portID, channelID)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, clientState)
	}
}

func queryNextSequenceRecvHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		sequenceRes, err := utils.QueryNextSequenceReceive(clientCtx, portID, channelID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(sequenceRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, sequenceRes)
	}
}
