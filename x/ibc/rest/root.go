package rest

import (
	"github.com/gorilla/mux"

	keys "github.com/tendermint/go-crypto/keys"

	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/ibc/{destchain}/{address}/send", TransferRequestHandler(cdc, kb)).Methods("POST")
}
