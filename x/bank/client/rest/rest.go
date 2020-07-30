package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// RegisterHandlers registers all x/bank transaction and query HTTP REST handlers
// on the provided mux router.
func RegisterHandlers(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/bank/accounts/{address}/transfers", NewSendRequestHandlerFn(clientCtx)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/bank/total", totalSupplyHandlerFn(clientCtx)).Methods("GET")
	r.HandleFunc("/bank/total/{denom}", supplyOfHandlerFn(clientCtx)).Methods("GET")
}
