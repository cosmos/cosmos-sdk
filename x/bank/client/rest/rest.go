package rest

import (
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
)

// RegisterRoutes registers bank-related REST handlers to a router
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	registerSendTxRoutes(cliCtx, r, cdc, kb)
}

// RegisterLiteRoutes registers bank REST handlers to gaia-lite
func RegisterLiteRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	registerQueryRoutes(cliCtx, r, cdc, "acc")
	registerSendTxRoutes(cliCtx, r, cdc, kb)
}
