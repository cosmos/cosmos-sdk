package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/client/utils"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc(fmt.Sprintf("/ibc/ports/{%s}/channels/{%s}/next-sequence-recv", RestPortID, RestChannelID), queryNextSequenceRecvHandlerFn(cliCtx)).Methods("GET")
}

// queryNextSequenceRecvHandlerFn implements a next sequence receive querying route
//
// @Summary Query next sequence receive
// @Tags IBC
// @Produce  json
// @Param port-id path string true "Port ID"
// @Param channel-id path string true "Channel ID"
// @Success 200 {object} QueryNextSequenceRecv "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid port id or channel id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/ports/{port-id}/channels/{channel-id}/next-sequence-recv [get]
func queryNextSequenceRecvHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portID := vars[RestPortID]
		channelID := vars[RestChannelID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		sequenceRes, err := utils.QueryNextSequenceRecv(cliCtx, portID, channelID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(sequenceRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, sequenceRes)
	}
}
