package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/rest"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	"github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/ibc/clients/localhost", createClientHandlerFn(cliCtx)).Methods("POST")
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
// @Router /ibc/clients/localhost [post]
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

		msg := types.NewMsgCreateClient(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, cliCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
