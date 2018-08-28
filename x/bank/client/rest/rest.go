package rest


import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gin-gonic/gin"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterSwaggerRoutes(routerGroup *gin.RouterGroup, ctx context.CLIContext, cdc *wire.Codec, kb keys.Keybase) {
	registerSwaggerQueryRoutes(routerGroup, ctx, cdc, "acc")
	registerSwaggerTxRoutes(routerGroup, ctx, cdc, kb)
}