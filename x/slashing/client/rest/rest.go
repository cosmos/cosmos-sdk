package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
)

func RegisterHandlers(ctx client.Context, r *mux.Router) {
	registerQueryRoutes(ctx, r)
	registerTxHandlers(ctx, r)
}

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx client.Context, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}
