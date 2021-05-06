package rest

import (
	"fmt"
	"net/http"

	"github.com/cosmos/cosmos-sdk/client"
	clientrest "github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
	"github.com/gorilla/mux"
)

const (
	MethodGet = "GET"
)

// RegisterRoutes registers poolx-related REST handlers to a router
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	r = clientrest.WithHTTPDeprecationHeaders(r)

	registerQueryRoutes(clientCtx, r)
	registerTxHandlers(clientCtx, r)
}

func registerQueryRoutes(clientCtx client.Context, r *mux.Router) {
	// Get the current poolx parameter values
	r.HandleFunc(
		"/poolx/parameters",
		paramsHandlerFn(clientCtx),
	).Methods("GET")
}

func registerTxHandlers(clientCtx client.Context, r *mux.Router) {
}

// HTTP request handler to query the distribution params values
func paramsHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientCtx, ok := rest.ParseQueryHeightOrReturnBadRequest(w, clientCtx, r)
		if !ok {
			return
		}

		route := fmt.Sprintf("custom/%s/%s", types.QuerierRoute, types.QueryParams)
		res, height, err := clientCtx.QueryWithData(route, nil)
		if rest.CheckInternalServerError(w, err) {
			return
		}

		clientCtx = clientCtx.WithHeight(height)
		rest.PostProcessResponse(w, clientCtx, res)
	}
}
