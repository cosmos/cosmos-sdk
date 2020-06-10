package rest

import (
	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	types "github.com/cosmos/cosmos-sdk/x/ibc/06-solomachine/types"
)

// REST client flags
const (
	RestClientID = "client-id"
)

// RegisterRoutes - General function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router, queryRoute string) {
	registerTxRoutes(clientCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq        rest.BaseReq         `json:"base_req" yaml:"base_req"`
	ClientID       string               `json:"client_id" yaml:"client_id"`
	ConsensusState types.ConsensusState `json:"consensus_state" yaml:"consensus_state"`
}

// UpdateClientReq defines the properties of an update client request's body.
type UpdateClientReq struct {
	BaseReq  rest.BaseReq `json:"base_req" yaml:"base_req"`
	ClientID string       `json:"client_id" yaml:"client_id"`
	Header   types.Header `json:"header" yaml:"header"`
}

// SubmitMisbehaviourReq defines the properties of a submit misbehaviour request's body.
type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq              `json:"base_req" yaml:"base_req"`
	Evidence evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
}
