package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

const (
	RestClientID   = "client-id"
	RestRootHeight = "height"
)

func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	registerQueryRoutes(cliCtx, r, queryRoute)
	registerTxRoutes(cliCtx, r)
}

type CreateClientReq struct {
	BaseReq        rest.BaseReq            `json:"base_req" yaml:"base_req"`
	ClientID       string                  `json:"client_id" yaml:"client_id"`
	ConsensusState exported.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

type UpdateClientReq struct {
	BaseReq  rest.BaseReq    `json:"base_req" yaml:"base_req"`
	ClientID string          `json:"client_id" yaml:"client_id"`
	Header   exported.Header `json:"header" yaml:"header"`
}

type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq      `json:"base_req" yaml:"base_req"`
	ClientID string            `json:"client_id" yaml:"client_id"`
	Evidence exported.Evidence `json:"evidence" yaml:"evidence"`
}
