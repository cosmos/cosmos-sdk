package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
)

// RegisterHandlers registers all x/bank transaction and query HTTP REST handlers
// on the provided mux router.
func RegisterHandlers(ctx context.CLIContext, m codec.Marshaler, txg tx.Generator, r *mux.Router) {
	r.HandleFunc("/bank/accounts/{address}/transfers", NewSendRequestHandlerFn(ctx, m, txg)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(ctx)).Methods("GET")
}

// ---------------------------------------------------------------------------
// Deprecated
//
// TODO: Remove once client-side Protobuf migration has been completed.
// ---------------------------------------------------------------------------

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	r.HandleFunc("/bank/accounts/{address}/transfers", SendRequestHandlerFn(cliCtx)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}", QueryBalancesRequestHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bank/total", totalSupplyHandlerFn(cliCtx)).Methods("GET")
	r.HandleFunc("/bank/total/{denom}", supplyOfHandlerFn(cliCtx)).Methods("GET")
}
