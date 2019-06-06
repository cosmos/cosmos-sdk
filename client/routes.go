package client

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
)

// Register routes
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *codec.Codec) {
	RegisterRPCRoutes(cliCtx, r)
	RegisterTxRoutes(cliCtx, r, cdc)
}
