package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/04-channel/client/utils"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}", RestPortID, RestChannelID), queryChannelHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/client_state", RestPortID, RestChannelID), queryChannelClientStateHandlerFn(clientCtx)).Methods("GET")
}

// queryChannelHandlerFn implements a channel querying route
//
// @Summary Query channel
// @Tags IBC
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryChannel "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id} [get]
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
