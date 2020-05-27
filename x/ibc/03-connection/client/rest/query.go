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
	r.HandleFunc("/ibc/connections", queryConnectionsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc("/ibc/clients/connections", queryClientsConnectionsHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(cliCtx, queryRoute)).Methods("GET")
}

// queryConnectionsHandlerFn implements connections querying route
//
// @Summary Query a client connection paths
// @Tags IBC
// @Produce json
// @Param page query int false "The page number to query" default(1)
// @Param limit query int false "The number of results per page" default(100)
// @Success 200 {object} QueryConnection "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/connections [get]
func queryClientsConnectionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		connections, height, err := utils.QueryAllConnections(cliCtx, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, connections)
	}
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
func queryConnectionHandlerFn(cliCtx context.CLIContext, _ string) http.HandlerFunc {
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

// queryConnectionsHandlerFn implements a client connections paths querying route
//
// @Summary Query all client connection paths
// @Tags IBC
// @Produce json
// @Param page query int false "The page number to query" default(1)
// @Param limit query int false "The number of results per page" default(100)
// @Success 200 {object} QueryClientsConnections "OK"
// @Failure 400 {object} rest.ErrorResponse "Bad Request"
// @Failure 500 {object} rest.ErrorResponse "Internal Server Error"
// @Router /ibc/clients/connections [get]
func queryConnectionsHandlerFn(cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, page, limit, err := rest.ParseHTTPArgsWithLimit(r, 0)
		if rest.CheckBadRequestError(w, err) {
			return
		}

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		connectionsPaths, height, err := utils.QueryAllClientConnectionPaths(cliCtx, page, limit)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(height)
		rest.PostProcessResponse(w, cliCtx, connectionsPaths)
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
func queryClientConnectionsHandlerFn(cliCtx context.CLIContext, _ string) http.HandlerFunc {
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
