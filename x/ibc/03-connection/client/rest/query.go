package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/client/utils"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
)

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/ibc/connections", queryConnectionsHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/connections/{%s}", RestConnectionID), queryConnectionHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc(fmt.Sprintf("/ibc/clients/{%s}/connections", RestClientID), queryClientConnectionsHandlerFn(clientCtx)).Methods("GET")
}

func queryConnectionsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		queryClient := types.NewQueryClient(clientCtx)

		req := &types.QueryConnectionsRequest{
			Pagination: &query.PageRequest{},
		}

		res, err := queryClient.Connections(context.Background(), req)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(res.Height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}

func queryConnectionHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionID := vars[RestConnectionID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		connRes, err := utils.QueryConnection(clientCtx, connectionID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(connRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, connRes)
	}
}

func queryClientConnectionsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clientID := vars[RestClientID]
		prove := rest.ParseQueryParamBool(r, flags.FlagProve)

		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		connPathsRes, err := utils.QueryClientConnections(clientCtx, clientID, prove)
		if err != nil {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			return
		}

		clientCtx = clientCtx.WithHeight(int64(connPathsRes.ProofHeight))
		rest.PostProcessResponse(w, clientCtx, connPathsRes)
	}
}
