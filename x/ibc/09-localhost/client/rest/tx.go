package rest

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client"
	sdk "github.com/KiraCore/cosmos-sdk/types"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	authclient "github.com/KiraCore/cosmos-sdk/x/auth/client"
	"github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/ibc/clients/localhost", createClientHandlerFn(clientCtx)).Methods("POST")
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
func createClientHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateClientReq
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

		msg := types.NewMsgCreateClient(fromAddr)
		if err := msg.ValidateBasic(); err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		authclient.WriteGenerateStdTxResponse(w, clientCtx, req.BaseReq, []sdk.Msg{msg})
	}
}
