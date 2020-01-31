package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(cliCtx, queryRoute)).Methods("GET")
}

// queryConnectionHandlerFn implements a connection querying route
//
// @Summary Query connection
// @Tags IBC
// @Produce  json
// @Param connection-id path string true "Client ID"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryConnection "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid connection id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections/{connection-id} [get]
func queryConnectionHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		connRes, err := utils.QueryConnection(cliCtx, connectionID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(connRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, connRes)
	}
}

// queryClientConnectionsHandlerFn implements a client connections querying route
//
// @Summary Query connections of a client
// @Tags IBC
// @Produce  json
// @Param client-id path string true "Client ID"
// @Param prove query boolean false "Proof of result"
// @Success 200 {object} QueryClientConnections "OK"
// @Failure 400 {object} rest.ErrorResponse "Invalid client id"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/{client-id}/connections [get]
func queryClientConnectionsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		connPathsRes, err := utils.QueryClientConnections(cliCtx, clientID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(int64(connPathsRes.ProofHeight))
		rest.PostProcessResponse(w, cliCtx, connPathsRes)
	}
}
