package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// REST query and parameter values
const (
	MethodGet = "GET"
)

// RegisterRoutes registers the auth module REST routes.
func RegisterRoutes(clientCtx client.Context, r *mux.Router, storeName string) {
	r.HandleFunc(
		"/auth/accounts/{address}", QueryAccountRequestHandlerFn(storeName, clientCtx),
	).Methods(MethodGet)

	r.HandleFunc(
		"/auth/params",
		queryParamsHandler(clientCtx),
	).Methods(MethodGet)
}

// RegisterTxRoutes registers all transaction routes on the provided router.
func RegisterTxRoutes(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/txs", QueryTxsRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/txs", BroadcastTxRequest(clientCtx)).Methods("POST")
	r.HandleFunc("/txs/encode", EncodeTxRequestHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc("/txs/decode", DecodeTxRequestHandlerFn(clientCtx)).Methods("POST")
}
