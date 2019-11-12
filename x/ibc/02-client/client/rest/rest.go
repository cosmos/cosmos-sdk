package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	RestClientID   = "client-id"
	RestRootHeight = "height"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

type CreateClientReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type UpdateClientReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type SubmitMisbehaviourReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}
