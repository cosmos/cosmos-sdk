package rest

import (
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/gorilla/mux"
)

func RegisterRoutes(r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc("/accounts/{address}/send", SendRequestHandler(cdc)).Methods("POST")
}
