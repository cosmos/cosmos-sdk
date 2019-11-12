package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	RestConnectionID = "connection-id"
	RestClientID     = "client-id"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

type ConnectionOpenInitReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ConnectionOpenTryReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ConnectionOpenAckReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ConnectionOpenConfirmReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}
