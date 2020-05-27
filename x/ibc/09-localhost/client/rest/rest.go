package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx client.Context, r *mux.Router, queryRoute string) {
	registerTxRoutes(cliCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}
