package rpc

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

// Register routes
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	RegisterRPCRoutes(cliCtx, r)
}
