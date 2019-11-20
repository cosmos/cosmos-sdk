package rest

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

const (
	Prove = "true"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(cliCtx, queryRoute)).Methods("GET")
}

func queryConnectionHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, err := cliCtx.QueryABCIWithProof(r, "ibc", types.PrefixKeyConnection(connectionID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var connection types.ConnectionEnd
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &connection); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, types.NewConnectionResponse(connectionID, connection, res.Proof, res.Height))
	}
}

func queryClientConnectionsHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		res, err := cliCtx.QueryABCIWithProof(r, "ibc", types.PrefixKeyClientConnections(clientID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		var paths []string
		if err := cliCtx.Codec.UnmarshalBinaryLengthPrefixed(res.Value, &paths); err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
		}
		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, types.NewClientConnectionsResponse(clientID, paths, res.Proof, res.Height))
	}
}
