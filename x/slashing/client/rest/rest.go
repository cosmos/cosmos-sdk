package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/gorilla/mux"
)

// RegisterRoutes registers staking-related REST handlers to a router
func RegisterRoutes(queryCtx context.QueryContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	registerQueryRoutes(queryCtx, r, cdc)
	registerTxRoutes(queryCtx, r, cdc, kb)
}
