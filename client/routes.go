package client

import (
	"github.com/gorilla/mux"

	"github.com/KiraCore/cosmos-sdk/client/context"
	"github.com/KiraCore/cosmos-sdk/client/rpc"
)

// Register routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	rpc.RegisterRPCRoutes(cliCtx, r)
}
