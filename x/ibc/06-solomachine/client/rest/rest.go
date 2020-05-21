package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	solomachinetypes "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

// REST client flags
const (
	RestClientID = "client-id"
)

// RegisterRoutes - General function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	registerTxRoutes(cliCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq        rest.BaseReq                    `json:"base_req" yaml:"base_req"`
	ClientID       string                          `json:"client_id" yaml:"client_id"`
	ConsensusState solomachinetypes.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

// UpdateClientReq defines the properties of an update client request's body.
type UpdateClientReq struct {
	BaseReq  rest.BaseReq            `json:"base_req" yaml:"base_req"`
	ClientID string                  `json:"client_id" yaml:"client_id"`
	Header   solomachinetypes.Header `json:"header" yaml:"header"`
}

// SubmitMisbehaviourReq defines the properties of a submit misbehaviour request's body.
type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq              `json:"base_req" yaml:"base_req"`
	Evidence evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
}
