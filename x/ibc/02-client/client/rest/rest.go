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

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	registerQueryRoutes(cliCtx, r, queryRoute)
	registerTxRoutes(cliCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq        rest.BaseReq            `json:"base_req" yaml:"base_req"`
	ClientID       string                  `json:"client_id" yaml:"client_id"`
	ConsensusState exported.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

// UpdateClientReq defines the properties of a update client request's body.
type UpdateClientReq struct {
	BaseReq rest.BaseReq    `json:"base_req" yaml:"base_req"`
	Header  exported.Header `json:"header" yaml:"header"`
}

// SubmitMisbehaviourReq defines the properties of a submit misbehaviour request's body.
type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq      `json:"base_req" yaml:"base_req"`
	Evidence exported.Evidence `json:"evidence" yaml:"evidence"`
}
