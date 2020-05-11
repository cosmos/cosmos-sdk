package client

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/rpc"
)

// Register routes
func RegisterRoutes(cliCtx context.Context, r *mux.Router) {
	rpc.RegisterRPCRoutes(cliCtx, r)
}
