package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
)

// RegisterRoutes registers bank-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
	r.HandleFunc("/tx/broadcast", BroadcastTxRequestHandlerFn(cdc, cliCtx)).Methods("POST")
}

// RegisterLiteRoutes registers bank REST handlers to gaia-lite
func RegisterLiteRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/bank/{address}/transfers", SendRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
	r.HandleFunc("/tx/broadcast", BroadcastTxRequestHandlerFn(cdc, cliCtx)).Methods("POST")
	r.HandleFunc("/bank/balances/{address}",
		QueryBalancesRequestHandlerFn("acc", cdc, authcmd.GetAccountDecoder(cdc), cliCtx),
	).Methods("GET")
}
