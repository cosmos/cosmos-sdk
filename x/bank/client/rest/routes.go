package rest

import (
	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandlerFn(cdc, kb, ctx)).Methods("POST")
	r.HandleFunc("/accounts/{destchain}/{address}/send", TransferRequestHandlerFn(cdc, kb, ctx)).Methods("POST")
}
