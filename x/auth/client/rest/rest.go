package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes registers the auth module REST routes.
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, storeName string) {
	r.HandleFunc(
		"/auth/accounts/{address}",
		QueryAccountRequestHandlerFn(storeName, context.GetAccountDecoder(cliCtx.Codec), cliCtx),
	).Methods("GET")

	// TODO: Change path or mount under x/bank if possible
	r.HandleFunc(
		"/bank/balances/{address}",
		QueryBalancesRequestHandlerFn(storeName, context.GetAccountDecoder(cliCtx.Codec), cliCtx),
	).Methods("GET")
}

// RegisterTxRoutes registers all transaction routes on the provided router.
func RegisterTxRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/txs/{hash}", QueryTxRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/txs", QueryTxsByTagsRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/txs", BroadcastTxRequest(cliCtx)).Methods("POST")
	r.HandleFunc("/txs/encode", EncodeTxRequestHandlerFn(cliCtx)).Methods("POST")
}
