package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/bank/accounts/{address}/transfers", SendRequestHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(cliCtx)).Methods("GET")
}
