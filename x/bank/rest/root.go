package rest

import (
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandler(cdc)).Methods("POST")
}
