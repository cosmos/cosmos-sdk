package rest

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	registerQueryRoutes(ctx, r, cdc)
	registerTxRoutes(ctx, r, cdc, kb)
}
