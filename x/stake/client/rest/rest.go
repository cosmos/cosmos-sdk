package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/gorilla/mux"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	registerQueryRoutes(cliCtx, r, cdc)
	registerTxRoutes(cliCtx, r, cdc, kb)
}

// RegisterSwaggerRoutes registers staking status query REST API handlers to a router
func RegisterSwaggerRoutes(routerGroup *gin.RouterGroup, ctx context.CLIContext, cdc *wire.Codec, kb keys.Keybase) {
	registerSwaggerQueryRoutes(routerGroup, ctx, cdc)
	registerSwaggerTxRoutes(routerGroup, ctx, cdc, kb)
}