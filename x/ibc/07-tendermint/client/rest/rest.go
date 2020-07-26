package rest

import (
	"time"

	ics23 "github.com/confio/ics23/go"
	"github.com/gorilla/mux"

	tmmath "github.com/tendermint/tendermint/libs/math"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/types/rest"
	evidenceexported "github.com/KiraCore/cosmos-sdk/x/evidence/exported"
	ibctmtypes "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
)

// REST client flags
const (
	RestClientID   = "client-id"
	RestRootHeight = "height"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(clientCtx client.Context, r *mux.Router) {
	registerTxRoutes(clientCtx, r)
}

// CreateClientReq defines the properties of a create client request's body.
type CreateClientReq struct {
	BaseReq         rest.BaseReq       `json:"base_req" yaml:"base_req"`
	ClientID        string             `json:"client_id" yaml:"client_id"`
	ChainID         string             `json:"chain_id" yaml:"chain_id"`
	Header          ibctmtypes.Header  `json:"header" yaml:"header"`
	TrustLevel      tmmath.Fraction    `json:"trust_level" yaml:"trust_level"`
	TrustingPeriod  time.Duration      `json:"trusting_period" yaml:"trusting_period"`
	UnbondingPeriod time.Duration      `json:"unbonding_period" yaml:"unbonding_period"`
	MaxClockDrift   time.Duration      `json:"max_clock_drift" yaml:"max_clock_drift"`
	ProofSpecs      []*ics23.ProofSpec `json:"proof_specs" yaml:"proof_specs"`
}

// UpdateClientReq defines the properties of a update client request's body.
type UpdateClientReq struct {
	BaseReq rest.BaseReq      `json:"base_req" yaml:"base_req"`
	Header  ibctmtypes.Header `json:"header" yaml:"header"`
}

// SubmitMisbehaviourReq defines the properties of a submit misbehaviour request's body.
type SubmitMisbehaviourReq struct {
	BaseReq  rest.BaseReq              `json:"base_req" yaml:"base_req"`
	Evidence evidenceexported.Evidence `json:"evidence" yaml:"evidence"`
}
