package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
)

func RegisterHandlers(ctx context.CLIContext, m codec.Marshaler, txg tx.Generator, r *mux.Router) {
	registerQueryRoutes(ctx, r)
	registerTxHandlers(ctx, m, txg, r)
}

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}
