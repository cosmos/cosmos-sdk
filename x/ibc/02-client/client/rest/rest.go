package rest

import (
	"time"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/types/rest"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// REST client flags
const (
	RestClientID   = "client-id"
	RestRootHeight = "height"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(cliCtx context.CLIContext, r *mux.Router, queryRoute string) {
	registerQueryRoutes(cliCtx, r)
	registerTxRoutes(cliCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq         rest.BaseReq            `json:"base_req" yaml:"base_req"`
	ClientID        string                  `json:"client_id" yaml:"client_id"`
	ConsensusState  exported.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
	TrustingPeriod  time.Duration           `json:"trusting_period" yaml:"trusting_period"`
	UnbondingPeriod time.Duration           `json:"unbonding_period" yaml:"unbonding_period"`
}

// UpdateClientReq defines the properties of a update client request's body.
type UpdateClientReq struct {
	BaseReq   rest.BaseReq    `json:"base_req" yaml:"base_req"`
	OldHeader exported.Header `json:"old_header" yaml:"old_header"`
	NewHeader exported.Header `json:"new_header" yaml:"new_header"`
}

// SubmitMisbehaviourReq defines the properties of a submit misbehaviour request's body.
type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq              `json:"base_req" yaml:"base_req"`
	Evidence evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
}
