package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
)

const (
	RestChannelID = "channel-id"
	RestPortID    = "port-id"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

type ChannelOpenInitReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ChannelOpenTryReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ChannelOpenAckReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ChannelOpenConfirmReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ChannelCloseInitReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}

type ChannelCloseConfirmReq struct {
	BaseReq rest.BaseReq `json:"base_req" yaml:"base_req"`
}
