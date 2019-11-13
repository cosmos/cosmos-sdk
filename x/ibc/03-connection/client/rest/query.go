package rest

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	r.HandleFunc(fmt.Sprintf("/ibc/connection/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(cliCtx, queryRoute)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/connection/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(cliCtx, queryRoute)).Methods("GET")
}

func queryConnectionHandlerFn(cliCtx context.CLIContext, queryRoute string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]

		cliCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, cliCtx, r)
		if !ok {
			return
		}

		// return proof if the prove query param is set to true
		proveStr := r.FormValue("prove")
		prove := false
		if strings.ToLower(strings.TrimSpace(proveStr)) == "true" {
			prove = true
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryConnectionParams(connectionID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req := abci.RequestQuery{
			Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryConnection),
			Data:  bz,
			Prove: prove,
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, res)
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

		// return proof if the prove query param is set to true
		proveStr := r.FormValue("prove")
		prove := false
		if strings.ToLower(strings.TrimSpace(proveStr)) == "true" {
			prove = true
		}

		bz, err := cliCtx.Codec.MarshalJSON(types.NewQueryClientConnectionsParams(clientID))
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
			return
		}

		req := abci.RequestQuery{
			Path:  fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryClientConnections),
			Data:  bz,
			Prove: prove,
		}

		res, err := cliCtx.QueryABCI(req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		cliCtx = cliCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, cliCtx, res)
	}
}
